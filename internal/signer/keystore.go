package signer

import (
	"fmt"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"

	"github.com/selesy/x402-buyer/pkg/api"
)

var _ api.EVMSigner = (*KeyStoreSigner)(nil)

type KeyStoreSigner struct {
	ks   *keystore.KeyStore
	acct accounts.Account
	pass []byte
}

func NewKeyStoreSigner(ks *keystore.KeyStore, acct accounts.Account, pass []byte) (*KeyStoreSigner, error) {
	if !ks.HasAddress(acct.Address) {
		return nil, fmt.Errorf("%w: %s", ErrAccountNotFound, acct.Address.Hex())
	}

	return &KeyStoreSigner{
		ks:   ks,
		acct: acct,
		pass: pass,
	}, nil
}

func (s *KeyStoreSigner) Address() common.Address {
	return s.acct.Address
}

func (s *KeyStoreSigner) Sign(digestHash []byte) ([]byte, error) {
	return s.ks.SignHashWithPassphrase(s.acct, string(s.pass), digestHash)
}
