package main

import (
	"os"
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

	log.Infof("Wallet Address: %s", os.Getenv("WALLET_ADDRESS"))
	log.Infof("Chain ID: %s", os.Getenv("CHAIN_ID"))
	log.Infof("Target Token: %s, Name: %s, Decimals: %s, Address: %s", os.Getenv("TARGET_TOKEN_SYMBOL"), os.Getenv("TARGET_TOKEN_NAME"), os.Getenv("TARGET_TOKEN_DECIMALS"), os.Getenv("TARGET_TOKEN_ADDRESS"))
	log.Infof("Stable Token: %s, Name: %s, Decimals: %s, Address: %s", os.Getenv("STABLE_TOKEN_SYMBOL"), os.Getenv("STABLE_TOKEN_NAME"), os.Getenv("STABLE_TOKEN_DECIMALS"), os.Getenv("STABLE_TOKEN_ADDRESS"))

	r := NewOneInchRouter(os.Getenv("ROUTER_CONTRACT_ADDRESS"))
	log.Infof("Router Contract Address: %s", r.RouterContractAddress())

	for {
		log.Debug("Generating access token...")
		if err := r.GenerateAccessToken(); err != nil {
			log.Fatalf("Error occurred while generating access token: %v, exiting...", err)
		}
		log.Debugf("Access Token: %s", r.AccessToken())
		log.Debugf("Expiration: %d", r.Expiration())
		log.Debug("Generated access token successfully")

		log.Debug("Generating swap quote...")
		quote, err := r.GetQuote(os.Getenv("WALLET_ADDRESS"), os.Getenv("CHAIN_ID"), os.Getenv("TARGET_TOKEN_ADDRESS"), os.Getenv("STABLE_TOKEN_ADDRESS"), "911447")
		if err != nil {
			log.Fatalf("Error occurred while generating quote: %v, exiting...", err)
		}
		log.Infof("Current Exchange Rate: %s => %s", quote.FromTokenAmount, quote.ToTokenAmount)
		log.Debug("Generated swap quote successfully")

		log.Debug("Checking pricing conditions...")

		log.Infof("Sleeping for 30 seconds before next request...")
		time.Sleep(30 * time.Second)
	}
}
