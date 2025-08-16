package signer

import (
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/selesy/x402-buyer/pkg/api"
)

var _ api.EVMSigner = (*ECDSASigner)(nil)

// ECDSASigner is an api.Signer that creates a cryptographic signature
// using an ecdsa.PrivateKey.
type ECDSASigner struct {
	priv *ecdsa.PrivateKey
}

func NewECDSASigner(priv *ecdsa.PrivateKey) (*ECDSASigner, error) {
	if priv.Curve != secp256k1.S256() {
		return nil, ErrInvalidCurve
	}

	if !secp256k1.S256().IsOnCurve(priv.X, priv.Y) {
		return nil, ErrInvalidPoint
	}

	return &ECDSASigner{
		priv: priv,
	}, nil
}

func NewECDSASignerFromBytes(b []byte) (*ECDSASigner, error) {
	priv := new(ecdsa.PrivateKey)
	priv.D = new(big.Int).SetBytes(b)
	priv.PublicKey.Curve = secp256k1.S256()
	priv.PublicKey.X, priv.PublicKey.Y = secp256k1.S256().ScalarBaseMult(priv.D.Bytes())

	if priv.X == nil || priv.Y == nil {
		return nil, ErrInvalidPoint
	}

	return NewECDSASigner(priv)
}

func NewECDSASignerFromHex(s string) (*ECDSASigner, error) {
	privBytes, err := hex.DecodeString(s)
	if err != nil {
		return nil, err
	}

	return NewECDSASignerFromBytes(privBytes)
}

func NewECDSASignerFromEnv(name string) (*ECDSASigner, error) {
	privHex := os.Getenv(name)
	if privHex == "" {
		return nil, fmt.Errorf("%w: %s", ErrEnvVarNotFound, name)
	}

	return NewECDSASignerFromHex(privHex)
}

func (s *ECDSASigner) Address() common.Address {
	return crypto.PubkeyToAddress(s.priv.PublicKey)
}

func (s *ECDSASigner) Sign(digestHash []byte) ([]byte, error) {
	return crypto.Sign(digestHash, s.priv)
}
