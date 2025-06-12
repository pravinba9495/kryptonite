package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

// QuoteResponse represents the response structure for a swap quote from the 1inch API.
type QuoteResponse struct {
	QuoteId           string `json:"quoteId"`
	FromTokenAmount   string `json:"fromTokenAmount"`
	ToTokenAmount     string `json:"toTokenAmount"`
	RecommendedPreset string `json:"recommended_preset"`
	Raw               string `json:"raw"`
}

type CreateOrderResponseMessageType struct {
	Maker        string `json:"maker"`
	MakerAsset   string `json:"makerAsset"`
	TakerAsset   string `json:"takerAsset"`
	MakerTraits  string `json:"makerTraits"`
	Salt         string `json:"salt"`
	MakingAmount string `json:"makingAmount"`
	TakingAmount string `json:"takingAmount"`
	Receiver     string `json:"receiver"`
}

// CreateOrderResponse represents the response structure for creating a swap order on the 1inch API.
type CreateOrderResponse struct {
	TypedData struct {
		PrimaryType string `json:"primaryType"`
		Types       struct {
			EIP712Domain []struct {
				Name string `json:"name"`
				Type string `json:"type"`
			} `json:"EIP712Domain"`
			Order []struct {
				Name string `json:"name"`
				Type string `json:"type"`
			} `json:"Order"`
		} `json:"types"`
		Domain struct {
			Name              string `json:"name"`
			Version           string `json:"version"`
			ChainId           int    `json:"chainId"`
			VerifyingContract string `json:"verifyingContract"`
		} `json:"domain"`
		Message CreateOrderResponseMessageType `json:"message"`
	} `json:"typedData"`
	OrderHash string `json:"orderHash"`
	Extension string `json:"extension"`
}

// SubmitOrderRequestPayload represents the payload structure for submitting a swap order on the 1inch API.
type SubmitOrderRequestPayload struct {
	Extension string                         `json:"extension"`
	QuoteId   string                         `json:"quoteId"`
	Signature string                         `json:"signature"`
	Order     CreateOrderResponseMessageType `json:"order"`
}

// SubmitOrderResponse represents the response structure for submitting a swap order on the 1inch API.
type SubmitOrderResponse struct {
}

// BalancesAndAllowancesResponse represents the response structure for token balances and allowances from the 1inch API.
type BalancesAndAllowancesResponse map[string]struct {
	Balance   string `json:"balance"`
	Allowance string `json:"allowance"`
}

// OneInchRouter defines the interface for interacting with the 1inch API.
type OneInchRouter interface {
	// GenerateOrRefreshAccessToken generates or refreshes the access token for the 1inch API.
	GenerateOrRefreshAccessToken() error

	// GetWalletTokenBalancesAndRouterAllowances retrieves the balances and allowances for the specified wallet address
	GetWalletTokenBalancesAndRouterAllowances(walletAddress string) (BalancesAndAllowancesResponse, error)

	// GetQuote retrieves a swap quote from the 1inch API.
	GetQuote(walletAddress string, fromTokenAddress string, toTokenAddress string, fromTokenAmount string) (*QuoteResponse, error)

	// CreateOrder creates a swap order on the 1inch API.
	CreateOrder(walletAddress string, fromTokenAddress string, toTokenAddress string, fromTokenAmount string, quote *QuoteResponse) (*CreateOrderResponse, error)

	// SubmitOrder submits a swap order to the 1inch API.
	SubmitOrder(signatureHex string, order *CreateOrderResponse, quote *QuoteResponse) error

	// AccessToken returns the current access token.
	AccessToken() string

	// Expiration returns the expiration time of the current access token.
	Expiration() int64

	// RouterContractAddress returns the contract address of the 1inch router.
	RouterContractAddress() string

	// ChainID returns the blockchain network ID (e.g., "1" for Ethereum mainnet).
	ChainID() string
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
	return r.chainId
}

// GetWalletTokenBalancesAndRouterAllowances retrieves the token balances and router allowances for the specified wallet address.
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

	var responseBody map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &responseBody); err != nil {
		return nil, err
	}

	responseBody["slippage"] = responseBody["k"]
	responseBody["autoSlippage"] = responseBody["autoK"]

	bytes, err := json.Marshal(responseBody)
	if err != nil {
		return nil, err
	}

	var quoteResponse QuoteResponse
	if err := json.Unmarshal(bodyBytes, &quoteResponse); err != nil {
		return nil, err
	}

	quoteResponse.Raw = string(bytes)

	return &quoteResponse, nil
}

// CreateOrder creates a swap order on the 1inch API using the provided wallet address, token addresses, and amount.
func (r *oneInchRouter) CreateOrder(walletAddress string, fromTokenAddress string, toTokenAddress string, fromTokenAmount string, quote *QuoteResponse) (*CreateOrderResponse, error) {
	if quote == nil {
		return nil, errors.New("invalid quote, cannot be nil")
	}

	url := fmt.Sprintf("https://proxy-app.1inch.io/v2.0/fusion/quoter/v2.0/%s/quote/build", r.chainId)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(quote.Raw)))
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()

	q.Add("walletAddress", walletAddress)
	q.Add("amount", fromTokenAmount)
	q.Add("fromTokenAddress", fromTokenAddress)
	q.Add("toTokenAddress", toTokenAddress)
	q.Add("preset", quote.RecommendedPreset)
	q.Add("source", "0xe26b9977") // TODO(praveen): no idea what this param is for, but it is probably needed by the API

	req.URL.RawQuery = q.Encode()

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", r.session.AccessToken))
	req.Header.Add("Content-Type", "application/json; charset=utf-8")

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, errors.New("request failed, status code: " + resp.Status)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var createOrderResponse CreateOrderResponse
	if err := json.Unmarshal(bodyBytes, &createOrderResponse); err != nil {
		return nil, err
	}

	return &createOrderResponse, nil
}

// SubmitOrder submits a swap order to the 1inch API.
func (r *oneInchRouter) SubmitOrder(signatureHex string, order *CreateOrderResponse, quote *QuoteResponse) error {
	if order == nil {
		return errors.New("invalid order, cannot be nil")
	}

	if quote == nil {
		return errors.New("invalid quote, cannot be nil")
	}

	url := fmt.Sprintf("https://proxy-app.1inch.io/v2.0/fusion/relayer/v2.0/%s/order/submit", r.chainId)

	payload := SubmitOrderRequestPayload{
		Extension: order.Extension,
		QuoteId:   quote.QuoteId,
		Signature: signatureHex,
		Order: CreateOrderResponseMessageType{
			Maker:        order.TypedData.Message.Maker,
			MakerAsset:   order.TypedData.Message.MakerAsset,
			TakerAsset:   order.TypedData.Message.TakerAsset,
			MakerTraits:  order.TypedData.Message.MakerTraits,
			Salt:         order.TypedData.Message.Salt,
			MakingAmount: order.TypedData.Message.MakingAmount,
			TakingAmount: order.TypedData.Message.TakingAmount,
			Receiver:     order.TypedData.Message.Receiver,
		},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", r.session.AccessToken))
	req.Header.Add("Content-Type", "application/json; charset=utf-8")

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return errors.New("request failed, status code: " + resp.Status)
	}

	return nil
}

// GenerateOrRefreshAccessToken generates or refreshes the access token for the 1inch API.
func (r *oneInchRouter) GenerateOrRefreshAccessToken() error {
	now := time.Now().Unix()
	diff := r.Expiration() - now - int64((10 * time.Minute).Seconds()) // with 10 minute buffer

	if diff > 0 {
		return nil
	}

	const url = "https://proxy-app.1inch.io/v2.0/auth/token"

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

// NewOneInchRouter creates a new instance of OneInchRouter with the specified contract address and blockchain id.
func NewOneInchRouter(contractAddress string, chainId string) OneInchRouter {
	return &oneInchRouter{
		routerContractAddress: contractAddress,
		chainId:               chainId,
	}
}
