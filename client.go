package buyer

import (
	"crypto/ecdsa"
	"net/http"

	"github.com/selesy/x402-buyer/internal/signer"
	"github.com/selesy/x402-buyer/pkg/api"
)

// ClientForPrivateKey returns an http.Client capable of making payments
// using cryptocurrency from Ethereum account associated with the provided
// ECDSA private key (which is expected to be using the Ethereum secp256k1
// curve.)
func ClientForPrivateKey(priv *ecdsa.PrivateKey, opts ...Option) (*http.Client, error) {
	signer, err := signer.NewECDSASigner(priv)
	if err != nil {
		return nil, err
	}

	return ClientForSigner(signer, opts...)
}

// ClientForPrivateKeyHex is like ClientForPrivateKey except that the
// private key is parsed from the provided hexadecimal string.
func ClientForPrivateKeyHex(privHex string, opts ...Option) (*http.Client, error) {
	signer, err := signer.NewECDSASignerFromHex(privHex)
	if err != nil {
		return nil, err
	}

	return ClientForSigner(signer, opts...)
}

// ClientForPrivateKeyHexFromEnv is like ClientForPrivateKeyHex except that
// hexadecimal string is read from the environment variable selected by name.
func ClientForPrivateKeyHexFromEnv(name string, opts ...Option) (*http.Client, error) {
	signer, err := signer.NewECDSASignerFromEnv(name)
	if err != nil {
		return nil, err
	}

	return ClientForSigner(signer, opts...)
}

// ClientForSigner returns an http.Client capable of making x402 paybments
// using the provided api.Signer.
func ClientForSigner(signer api.Signer, opts ...Option) (*http.Client, error) {
	cfg, err := newConfig(opts...)
	if err != nil {
		return nil, err
	}

	cfg.client.Transport = newTransport(cfg.client.Transport, signer, cfg)

	return cfg.client, nil
}
