package main

import (
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
)

// Wallet interface defines methods for signing messages and retrieving the wallet address.
type Wallet interface {
	// SignEIP712Message signs an EIP-712 typed data message using the wallet's private key.
	SignEIP712Message(message []byte) ([]byte, error)

	// Address returns the wallet's address.
	Address() string

	// ChainID returns the blockchain network ID associated with the wallet.
	ChainID() string
}

// wallet implements the Wallet interface.
type wallet struct {
	// privateKey is the ECDSA private key used for signing messages.
	privateKey *ecdsa.PrivateKey

	// publicKey is the ECDSA public key derived from the private key.
	publicKey *ecdsa.PublicKey

	// address is the Ethereum address derived from the public key.
	address string

	// chainId is the blockchain network ID associated with the wallet.
	chainId string
}

// SignEIP712Message signs an EIP-712 typed data message using the wallet's private key.
func (w *wallet) SignEIP712Message(message []byte) ([]byte, error) {
	var typedData apitypes.TypedData
	err := json.Unmarshal(message, &typedData)
	if err != nil {
		return []byte(""), err
	}

	domainSeparator, err := typedData.HashStruct("EIP712Domain", typedData.Domain.Map())
	if err != nil {
		return []byte(""), err
	}

	typedDataHash, err := typedData.HashStruct(typedData.PrimaryType, typedData.Message)
	if err != nil {
		return []byte(""), err
	}

	rawData := []byte(fmt.Sprintf("\x19\x01%s%s", string(domainSeparator), string(typedDataHash)))
	digestHash := crypto.Keccak256(rawData)

	signature, err := crypto.Sign(digestHash, w.privateKey)
	if err != nil {
		return []byte(""), err
	}

	recoveredPubKey, err := crypto.Ecrecover(digestHash, signature)
	if err != nil {
		return []byte(""), err
	}

	publicKey, err := crypto.UnmarshalPubkey(recoveredPubKey)
	if err != nil {
		return []byte(""), err
	}

	recoveredAddr := crypto.PubkeyToAddress(*publicKey)
	if recoveredAddr.Hex() != w.address {
		return []byte(""), errors.New("signature does not match the wallet address")
	}

	if signature[64] < 27 {
		signature[64] += 27
	}

	return signature, nil
}

// Address returns the wallet's address.
func (w *wallet) Address() string {
	return w.address
}

// ChainID returns the blockchain network ID associated with the wallet.
func (w *wallet) ChainID() string {
	return w.chainId
}

// NewWallet creates a new Wallet instance from a given private key and expected address.
func NewWallet(privateKeyHex, expectedAddress string, chainId string) (Wallet, error) {
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return nil, err
	}

	publicKey := privateKey.PublicKey

	address := crypto.PubkeyToAddress(publicKey).Hex()

	if address != expectedAddress {
		return nil, errors.New("provided private key does not match the expected address")
	}

	return &wallet{
		privateKey: privateKey,
		publicKey:  &publicKey,
		address:    address,
		chainId:    chainId,
	}, nil
}
