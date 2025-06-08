package main

import (
	"bytes"
	"crypto/ecdsa"
	"errors"

	"github.com/ethereum/go-ethereum/crypto"
)

// Wallet interface defines methods for signing messages and retrieving the wallet address.
type Wallet interface {
	// SignMessage signs a message using the wallet's private key and returns the signature.
	SignMessage(message string) ([]byte, error)

	// VerifySignature verifies a message signature using the wallet's public key.
	VerifySignature(signature []byte, message string) error

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

// SignMessage signs a message using the wallet's private key and returns the signature.
func (w *wallet) SignMessage(message string) ([]byte, error) {
	hash := crypto.Keccak256Hash([]byte(message))

	signature, err := crypto.Sign(hash.Bytes(), w.privateKey)
	if err != nil {
		return []byte(""), err
	}

	return signature, nil
}

// Verify verifies a message signature using the wallet's private key
func (w *wallet) VerifySignature(signature []byte, message string) error {
	hash := crypto.Keccak256Hash([]byte(message))

	sigPublicKey, err := crypto.Ecrecover(hash.Bytes(), signature)
	if err != nil {
		return err
	}

	isMatching := bytes.Equal(sigPublicKey, crypto.FromECDSAPub(w.publicKey))

	if !isMatching {
		return errors.New("signature does not match the public key")
	}
	return nil
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
