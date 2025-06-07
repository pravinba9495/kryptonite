package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// QuoteResponse represents the response structure for a swap quote from the 1inch API.
type QuoteResponse struct {
	FromTokenAmount    string  `json:"fromTokenAmount"`
	ToTokenAmount      string  `json:"toTokenAmount"`
	K                  float64 `json:"k"`
	AutoK              float64 `json:"autoK"`
	MxK                float64 `json:"mxK"`
	IntegratorFee      float64 `json:"integratorFee"`
	MarketAmount       string  `json:"marketAmount"`
	FeeToken           string  `json:"feeToken"`
	Gas                int64   `json:"gas"`
	PfGas              int64   `json:"pfGas"`
	PriceImpactPercent float64 `json:"priceImpactPercent"`
	RecommendedPreset  string  `json:"recommended_preset"`
	SettlementAddress  string  `json:"settlementAddress"`
	Suggested          bool    `json:"suggested"`
	SurplusFee         float64 `json:"surplusFee"`

	Raw string
}

type BalancesAndAllowancesResponse map[string]struct {
	Balance   string `json:"balance"`
	Allowance string `json:"allowance"`
}

// OneInchRouter defines the interface for interacting with the 1inch API.
type OneInchRouter interface {
	// GenerateAccessToken generates a new access token for the 1inch API.
	GenerateAccessToken() error

	// GetQuote retrieves a swap quote from the 1inch API.
	GetQuote(walletAddress string, fromTokenAddress string, toTokenAddress string, fromTokenAmount string) (*QuoteResponse, error)

	// AccessToken returns the current access token.
	AccessToken() string

	// Expiration returns the expiration time of the current access token.
	Expiration() int64

	// RouterContractAddress returns the contract address of the 1inch router.
	RouterContractAddress() string

	// ChainID returns the blockchain network ID (e.g., "1" for Ethereum mainnet).
	ChainID() string

	// GetWalletTokenBalancesAndRouterAllowances retrieves the balances and allowances for the specified wallet address
	GetWalletTokenBalancesAndRouterAllowances(walletAddress string) (BalancesAndAllowancesResponse, error)
}

// oneInchRouterSession holds the access token and expiration time for the 1inch API.
type oneInchRouterSession struct {
	// AccessToken is the access token for the 1inch API.
	AccessToken string `json:"access_token"`

	// Exp is the expiration time of the access token in Unix timestamp format.
	Exp int64 `json:"exp"`
}

// oneInchRouter implements the OneInchRouter interface for interacting with the 1inch API.
type oneInchRouter struct {
	// session holds the access token and expiration time for the 1inch API.
	session *oneInchRouterSession

	// routerContractAddress is the contract address of the 1inch router.
	routerContractAddress string

	// chainId is the blockchain network ID (e.g., "1" for Ethereum mainnet).
	chainId string
}

// RouterContractAddress returns the contract address of the 1inch router.
func (r *oneInchRouter) RouterContractAddress() string {
	return r.routerContractAddress
}

// ChainID returns the blockchain network ID (e.g., "1" for Ethereum mainnet).
func (r *oneInchRouter) ChainID() string {
	return r.routerContractAddress
}

func (r *oneInchRouter) GetWalletTokenBalancesAndRouterAllowances(walletAddress string) (BalancesAndAllowancesResponse, error) {
	url := fmt.Sprintf("https://proxy-app.1inch.io/v2.0/balance/v1.2/%s/allowancesAndBalances/%s/%s", r.chainId, r.routerContractAddress, walletAddress)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", r.session.AccessToken))

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("request failed, status code: " + resp.Status)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var balancesAndAllowancesResponse BalancesAndAllowancesResponse
	if err := json.Unmarshal(bodyBytes, &balancesAndAllowancesResponse); err != nil {
		return nil, err
	}

	return balancesAndAllowancesResponse, nil
}

// GetQuote retrieves a swap quote from the 1inch API using the provided token addresses and amount.
func (r *oneInchRouter) GetQuote(walletAddress string, fromTokenAddress string, toTokenAddress string, fromTokenAmount string) (*QuoteResponse, error) {
	url := fmt.Sprintf("https://proxy-app.1inch.io/v2.0/fusion/quoter/v2.0/%s/quote/receive", r.chainId)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()

	q.Add("walletAddress", walletAddress)
	q.Add("amount", fromTokenAmount)
	q.Add("fromTokenAddress", fromTokenAddress)
	q.Add("toTokenAddress", toTokenAddress)

	q.Add("enableEstimate", "true")
	q.Add("showDestAmountMinusFee", "true")
	q.Add("source", "0xe26b9977") // TODO(praveen): no idea what this param is for, but it is probably needed by the API

	req.URL.RawQuery = q.Encode()

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", r.session.AccessToken))

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("request failed, status code: " + resp.Status)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var quoteResponse QuoteResponse
	if err := json.Unmarshal(bodyBytes, &quoteResponse); err != nil {
		return nil, err
	}
	quoteResponse.Raw = string(bodyBytes) // Store the raw response

	return &quoteResponse, nil
}

// GenerateAccessToken generates a new access token for the 1inch API.
func (r *oneInchRouter) GenerateAccessToken() error {
	const url = "https://proxy-app.1inch.io/v2.0/auth/token?ngsw-bypass"

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("request failed, status code: " + resp.Status)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(bodyBytes, &r.session); err != nil {
		return err
	}

	return nil
}

// AccessToken returns the current access token for the 1inch API.
func (r *oneInchRouter) AccessToken() string {
	if r.session == nil {
		return ""
	}
	return r.session.AccessToken
}

// Expiration returns the expiration time of the current access token in Unix timestamp format.
func (r *oneInchRouter) Expiration() int64 {
	if r.session == nil {
		return 0
	}
	return r.session.Exp
}

// NewOneInchRouter creates a new instance of OneInchRouter with the specified contract address.
func NewOneInchRouter(contractAddress string, chainId string) OneInchRouter {
	return &oneInchRouter{
		routerContractAddress: contractAddress,
		chainId:               chainId,
	}
}
