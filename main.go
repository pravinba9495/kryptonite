package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strconv"
	"time"

	"github.com/charmbracelet/log"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

func main() {
	log.SetLevel(log.DebugLevel)
	log.SetReportCaller(true)
	log.Info("Starting service...")

	if err := godotenv.Load(); err != nil {
		log.Warnf("Error occuured while loading .env file: %v", err)
	}

	env := os.Getenv("ENV")
	if env == "production" {
		log.SetLevel(log.InfoLevel)
	}

	walletExpectedAddress := os.Getenv("WALLET_ADDRESS")
	privateKeyHex := os.Getenv("WALLET_PRIVATE_KEY_HEX")
	chainId := os.Getenv("CHAIN_ID")
	targetTokenSymbol := os.Getenv("TARGET_TOKEN_SYMBOL")
	targetTokenName := os.Getenv("TARGET_TOKEN_NAME")
	targetTokenDecimals := os.Getenv("TARGET_TOKEN_DECIMALS")
	targetTokenAddress := os.Getenv("TARGET_TOKEN_ADDRESS")
	stableTokenSymbol := os.Getenv("STABLE_TOKEN_SYMBOL")
	stableTokenName := os.Getenv("STABLE_TOKEN_NAME")
	stableTokenDecimals := os.Getenv("STABLE_TOKEN_DECIMALS")
	stableTokenAddress := os.Getenv("STABLE_TOKEN_ADDRESS")
	routerContractAddress := os.Getenv("ROUTER_CONTRACT_ADDRESS")
	redisHost := os.Getenv("REDIS_HOST")
	redisPort := os.Getenv("REDIS_PORT")
	redisPassword := os.Getenv("REDIS_PASSWORD")

	log.Info("Connecting to redis...")
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisHost + ":" + redisPort,
		Password: redisPassword,
		DB:       0,
	})
	if err := rdb.Ping(context.TODO()).Err(); err != nil {
		log.Fatalf("Error occurred while connecting to Redis: %v, exiting...", err)
	}
	log.Info("Connected to redis successfully")

	w, err := NewWallet(privateKeyHex, walletExpectedAddress, chainId)
	if err != nil {
		log.Fatalf("Error occurred while creating wallet: %v, exiting...", err)
	}

	log.Infof("Wallet Address: %s", w.Address())
	log.Infof("Chain ID: %s", chainId)
	log.Infof("Target Token: %s, Name: %s, Decimals: %s, Address: %s", targetTokenSymbol, targetTokenName, targetTokenDecimals, targetTokenAddress)
	log.Infof("Stable Token: %s, Name: %s, Decimals: %s, Address: %s", stableTokenSymbol, stableTokenName, stableTokenDecimals, stableTokenAddress)

	r := NewOneInchRouter(routerContractAddress, chainId)
	log.Infof("Router Contract Address: %s", r.RouterContractAddress())
	log.Infof("Router Chain ID: %s", r.ChainID())

	pm := NewPriceMonitor(BuyOrder, 0, 0, 0.5, 1.0)
	done := false

	for {
		if err := r.GenerateOrRefreshAccessToken(); err != nil {
			log.Fatalf("Error occurred while generating/refreshing access token: %v, exiting...", err)
		}
		log.Debug("Generated/Refreshed access token successfully")
		log.Debugf("Access Token: %s", r.AccessToken())
		log.Debugf("Expiration: %d", r.Expiration())

		log.Debug("Fetching wallet token balances and router allowances...")
		balancesAndAllowances, err := r.GetWalletTokenBalancesAndRouterAllowances(w.Address())
		if err != nil {
			log.Fatalf("Error occurred while fetching token balances and router allowances: %v, exiting...", err)
		}
		log.Debugf("%s Balance: %s", targetTokenSymbol, balancesAndAllowances[targetTokenAddress].Balance)
		log.Debugf("%s Balance: %s", stableTokenSymbol, balancesAndAllowances[stableTokenAddress].Balance)
		log.Debugf("%s Allowance: %s", targetTokenSymbol, balancesAndAllowances[targetTokenAddress].Allowance)
		log.Debugf("%s Allowance: %s", stableTokenSymbol, balancesAndAllowances[stableTokenAddress].Allowance)
		log.Debug("Fetched wallet token balances and router allowances successfully")

		log.Debug("Checking router allowances...")
		if balancesAndAllowances[targetTokenAddress].Allowance == "0" {
			log.Fatalf("Insufficient router allowance for %s, exiting...", targetTokenSymbol)
		}
		if balancesAndAllowances[stableTokenAddress].Allowance == "0" {
			log.Fatalf("Insufficient router allowance for %s, exiting...", stableTokenSymbol)
		}
		log.Debug("Checked router allowances successfully")

		log.Debug("Checking token balances...")
		if balancesAndAllowances[targetTokenAddress].Balance == "0" && balancesAndAllowances[stableTokenAddress].Balance == "0" {
			log.Fatalf("Insufficient wallet balances for %s and %s, exiting...", targetTokenSymbol, stableTokenSymbol)
		}
		log.Debug("Checked token balances successfully")

		log.Debug("Updating token balances in redis...")
		b1, err := rdb.Get(context.TODO(), fmt.Sprintf("LAST_BALANCE:%s", targetTokenSymbol)).Result()
		if err == redis.Nil {
			log.Warnf("Key LAST_BALANCE:%s does not exist in Redis, creating it...", targetTokenSymbol)
			if err := rdb.Set(context.TODO(), fmt.Sprintf("LAST_BALANCE:%s", targetTokenSymbol), balancesAndAllowances[targetTokenAddress].Balance, 0).Err(); err != nil {
				log.Fatalf("Error occurred while setting LAST_BALANCE:%s in Redis: %v, exiting...", targetTokenSymbol, err)
			}
		} else {
			if err != nil {
				log.Fatalf("Error occurred while getting LAST_BALANCE:%s from Redis: %v, exiting...", targetTokenSymbol, err)
			}
		}

		if balancesAndAllowances[targetTokenAddress].Balance != "0" {
			if b1 != balancesAndAllowances[targetTokenAddress].Balance {
				if err := rdb.Set(context.TODO(), fmt.Sprintf("LAST_BALANCE:%s", targetTokenSymbol), balancesAndAllowances[targetTokenAddress].Balance, 0).Err(); err != nil {
					log.Fatalf("Error occurred while setting LAST_BALANCE:%s in Redis: %v, exiting...", targetTokenSymbol, err)
				}
				if err := rdb.LPush(context.TODO(), fmt.Sprintf("BALANCES:%s", targetTokenSymbol), balancesAndAllowances[targetTokenAddress].Balance).Err(); err != nil {
					log.Fatalf("Error occurred while pushing %s balance to Redis: %v, exiting...", targetTokenSymbol, err)
				}
			}
		}

		b2, err := rdb.Get(context.TODO(), fmt.Sprintf("LAST_BALANCE:%s", stableTokenSymbol)).Result()
		if err == redis.Nil {
			log.Warnf("Key LAST_BALANCE:%s does not exist in Redis, creating it...", stableTokenSymbol)
			if err := rdb.Set(context.TODO(), fmt.Sprintf("LAST_BALANCE:%s", stableTokenSymbol), balancesAndAllowances[stableTokenAddress].Balance, 0).Err(); err != nil {
				log.Fatalf("Error occurred while setting LAST_BALANCE:%s in Redis: %v, exiting...", stableTokenSymbol, err)
			}
		} else {
			if err != nil {
				log.Fatalf("Error occurred while getting LAST_BALANCE:%s from Redis: %v, exiting...", stableTokenSymbol, err)
			}
		}

		if balancesAndAllowances[stableTokenAddress].Balance != "0" {
			if b2 != balancesAndAllowances[stableTokenAddress].Balance {
				if err := rdb.Set(context.TODO(), fmt.Sprintf("LAST_BALANCE:%s", stableTokenSymbol), balancesAndAllowances[stableTokenAddress].Balance, 0).Err(); err != nil {
					log.Fatalf("Error occurred while setting LAST_BALANCE:%s in Redis: %v, exiting...", stableTokenSymbol, err)
				}
				if err := rdb.LPush(context.TODO(), fmt.Sprintf("BALANCES:%s", stableTokenSymbol), balancesAndAllowances[stableTokenAddress].Balance).Err(); err != nil {
					log.Fatalf("Error occurred while pushing %s balance to Redis: %v, exiting...", stableTokenSymbol, err)
				}
			}
		}
		log.Debug("Updated token balances in redis successfully")

		var fromTokenAddress string
		var fromTokenSymbol string
		var fromTokenAmount string
		var fromTokenDecimals int
		var toTokenAddress string
		var toTokenSymbol string
		var toTokenDecimals int

		if balancesAndAllowances[targetTokenAddress].Balance == "0" && balancesAndAllowances[stableTokenAddress].Balance != "0" {
			done = true
			fromTokenAddress = stableTokenAddress
			fromTokenSymbol = stableTokenSymbol
			fromTokenAmount = balancesAndAllowances[stableTokenAddress].Balance
			fromTokenDecimals, err = strconv.Atoi(stableTokenDecimals)
			if err != nil {
				log.Fatalf("Error converting stable token decimals: %v, exiting...", err)
			}
			toTokenAddress = targetTokenAddress
			toTokenSymbol = targetTokenSymbol
			toTokenDecimals, err = strconv.Atoi(targetTokenDecimals)
			if err != nil {
				log.Fatalf("Error converting target token decimals: %v, exiting...", err)
			}
		}

		if balancesAndAllowances[targetTokenAddress].Balance != "0" && balancesAndAllowances[stableTokenAddress].Balance == "0" {
			if !done {
				pm.SwitchOrderType(SellOrder, 0, 0)
			}
			done = true
			fromTokenAddress = targetTokenAddress
			fromTokenSymbol = targetTokenSymbol
			fromTokenAmount = balancesAndAllowances[targetTokenAddress].Balance
			fromTokenDecimals, err = strconv.Atoi(targetTokenDecimals)
			if err != nil {
				log.Fatalf("Error converting target token decimals: %v, exiting...", err)
			}
			toTokenAddress = stableTokenAddress
			toTokenSymbol = stableTokenSymbol
			toTokenDecimals, err = strconv.Atoi(stableTokenDecimals)
			if err != nil {
				log.Fatalf("Error converting stable token decimals: %v, exiting...", err)
			}
		}

		log.Debugf("Waiting to swap from %s to %s, generating quote...", fromTokenSymbol, toTokenSymbol)
		quote, err := r.GetQuote(w.Address(), fromTokenAddress, toTokenAddress, fromTokenAmount)
		if err != nil {
			log.Fatalf("Error occurred while generating quote: %v, exiting...", err)
		}

		fromTokenAmountFloat, err := strconv.ParseFloat(fromTokenAmount, 64)
		if err != nil {
			log.Fatalf("Error converting fromTokenAmount to float: %v, exiting...", err)
		}

		quoteToTokenAmountFloat, err := strconv.ParseFloat(quote.ToTokenAmount, 64)
		if err != nil {
			log.Fatalf("Error converting toTokenAmount to float: %v, exiting...", err)
		}

		f1 := (fromTokenAmountFloat / math.Pow(10, float64(fromTokenDecimals)))
		f2 := (quoteToTokenAmountFloat / math.Pow(10, float64(toTokenDecimals)))
		currentPrice := 0.0

		if done {
			if pm.currentOrderType == BuyOrder {
				currentPrice = f1 / f2
			}

			if pm.currentOrderType == SellOrder {
				currentPrice = f2 / f1
			}
			pm.Update(currentPrice)
		}

		log.Infof("Current Exchange Rate: %f %s => %f %s", f1, fromTokenSymbol, f2, toTokenSymbol)
		log.Debug("Generated swap quote successfully")

		log.Debug("Creating order data...")
		order, err := r.CreateOrder(w.Address(), fromTokenAddress, toTokenAddress, fromTokenAmount, quote)
		if err != nil {
			log.Fatalf("Error occurred while creating order data for signing: %v, exiting...", err)
		}
		log.Debugf("Created order with hash: %s successfully", order.OrderHash)

		log.Debug("Signing order...")
		orderTypedDataBytes, err := json.Marshal(order.TypedData)
		if err != nil {
			log.Fatalf("Error occurred while marshaling order typed data: %v, exiting...", err)
		}
		signature, err := w.SignEIP712Message(orderTypedDataBytes)
		if err != nil {
			log.Fatalf("Error occurred while signing order: %v, exiting...", err)
		}
		signatureHex := hexutil.Encode(signature)
		log.Debugf("Signed EIP-712 Message Hex: %s", signatureHex)
		log.Debug("Signed order successfully")

		isTriggered := pm.isTriggered
		log.Infof("Waiting to %s, Triggered: %t, Current Price: 1 %s = %f %s, Up: %f %s, Down %f %s", pm.currentOrderType.String(), isTriggered, targetTokenSymbol, currentPrice, stableTokenSymbol, pm.triggerPriceUp, stableTokenSymbol, pm.triggerPriceDown, stableTokenSymbol)

		dur := 10 * time.Second
		if isTriggered && f2 > 971.00 {
			log.Info("Submitting order...")
			if err := r.SubmitOrder(signatureHex, order, quote); err != nil {
				log.Errorf("Error occurred while submitting order: %v", err)
			} else {
				log.Info("Order submitted successfully")
			}
			dur = 1 * time.Hour
		}
		log.Infof("Sleeping for %s before next request...", dur)
		time.Sleep(dur)
	}
}
