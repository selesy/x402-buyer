package decred

import (
	"encoding/hex"
	"net/http"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"

	buyer "github.com/selesy/x402-buyer"
	"github.com/selesy/x402-buyer/internal/signer"
)

// ClientForPrivateKey returns an http.Client capable of making payments
// using cryptocurrency from the Ethereum account associated with the provided
// ECDSA private key (which is expected to be using the Decred secp256k1
// curve.)
func ClientForPrivateKey(priv *secp256k1.PrivateKey, opts ...buyer.Option) (*http.Client, error) {
	signer, err := signer.NewECDSASigner(priv.ToECDSA())
	if err != nil {
		return nil, err
	}

	return buyer.ClientForSigner(signer, opts...)
}

// ClientForPrivateKeyHex is like ClientForPrivateKey except that the
// private key is parsed from the provided hexadecimal string.
func ClientForPrivateKeyHex(privHex string, opts ...buyer.Option) (*http.Client, error) {
	privBytes, err := hex.DecodeString(privHex)
	if err != nil {
		return nil, err
	}

	priv := secp256k1.PrivKeyFromBytes(privBytes)

	signer, err := signer.NewECDSASigner(priv.ToECDSA())
	if err != nil {
		return nil, err
	}

	return buyer.ClientForSigner(signer, opts...)
}

// ClientForPrivateKeyHexFromEnv is like ClientForPrivateKeyHex except that
// hexadecimal string is read from the environment variable selected by name.
func ClientForPrivateKeyHexFromEnv(name string, opts ...buyer.Option) (*http.Client, error) {
	signer, err := signer.NewECDSASignerFromEnv(name)
	if err != nil {
		return nil, err
	}

	return buyer.ClientForSigner(signer, opts...)
}
