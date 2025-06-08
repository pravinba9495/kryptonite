package main

import (
	"crypto/ecdsa"
	"errors"

	"github.com/ethereum/go-ethereum/crypto"
)

// Wallet interface defines methods for signing messages and retrieving the wallet address.
type Wallet interface {
	Sign(message string) (string, error)
	Address() string
	ChainID() string
}

// wallet implements the Wallet interface.
type wallet struct {
	privateKey *ecdsa.PrivateKey
	publicKey  *ecdsa.PublicKey
	address    string
	chainId    string
}

// Sign simulates signing a message with the wallet's private key.
func (w *wallet) Sign(message string) (string, error) {
	return message, nil
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
