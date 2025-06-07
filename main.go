package main

import (
	"math"
	"os"
	"strconv"
	"time"

	"github.com/charmbracelet/log"
	"github.com/joho/godotenv"
)

func main() {

	log.SetReportCaller(true)
	log.Info("Starting service...")

	if err := godotenv.Load(); err != nil {
		log.Warnf("Could not load .env file: %v", err)
	}

	log.SetLevel(log.DebugLevel)

	walletAddress := os.Getenv("WALLET_ADDRESS")
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

	log.Infof("Wallet Address: %s", walletAddress)
	log.Infof("Chain ID: %s", chainId)
	log.Infof("Target Token: %s, Name: %s, Decimals: %s, Address: %s", targetTokenSymbol, targetTokenName, targetTokenDecimals, targetTokenAddress)
	log.Infof("Stable Token: %s, Name: %s, Decimals: %s, Address: %s", stableTokenSymbol, stableTokenName, stableTokenDecimals, stableTokenAddress)

	r := NewOneInchRouter(routerContractAddress, chainId)
	log.Infof("Router Contract Address: %s", r.RouterContractAddress())
	log.Infof("Router Chain ID: %s", r.ChainID())

	for {
		log.Debug("Generating access token...")
		if err := r.GenerateAccessToken(); err != nil {
			log.Fatalf("Error occurred while generating access token: %v, exiting...", err)
		}
		log.Debugf("Access Token: %s", r.AccessToken())
		log.Debugf("Expiration: %d", r.Expiration())
		log.Debug("Generated access token successfully")

		log.Debug("Fetching wallet token balances and router allowances...")
		balancesAndAllowances, err := r.GetWalletTokenBalancesAndRouterAllowances(walletAddress)
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
			log.Fatalf("Insufficient router allowance for %s, creating approval tx...", targetTokenSymbol)
		}
		if balancesAndAllowances[stableTokenAddress].Allowance == "0" {
			log.Fatalf("Insufficient router allowance for %s, creating approval tx...", stableTokenSymbol)
		}
		log.Debug("Checked router allowances successfully")

		log.Debug("Checking token balances...")
		if balancesAndAllowances[targetTokenAddress].Balance == "0" && balancesAndAllowances[stableTokenAddress].Balance == "0" {
			log.Fatalf("Insufficient wallet balances for %s and %s, skipping...", targetTokenSymbol, stableTokenSymbol)
		}
		log.Debug("Checked token balances successfully")

		var fromTokenAddress string
		var fromTokenSymbol string
		var fromTokenAmount string
		var fromTokenDecimals int
		var toTokenAddress string
		var toTokenSymbol string
		var toTokenDecimals int

		if balancesAndAllowances[targetTokenAddress].Balance == "0" && balancesAndAllowances[stableTokenAddress].Balance != "0" {
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
		quote, err := r.GetQuote(walletAddress, fromTokenAddress, toTokenAddress, fromTokenAmount)
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

		log.Infof("Current Exchange Rate: %f %s => %f %s", (fromTokenAmountFloat / math.Pow(10, float64(fromTokenDecimals))), fromTokenSymbol, (quoteToTokenAmountFloat / math.Pow(10, float64(toTokenDecimals))), toTokenSymbol)
		log.Debug("Generated swap quote successfully")

		log.Debug("Checking pricing conditions...")

		log.Infof("Sleeping for 30 seconds before next request...")
		time.Sleep(30 * time.Second)
	}
}
