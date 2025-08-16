package signer_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/selesy/x402-buyer/internal/signer"
	"github.com/selesy/x402-buyer/pkg/api/apitest"
)

func TestECDSASigner(t *testing.T) {
	t.Parallel()

	t.Run("passes", func(t *testing.T) {
		t.Parallel()

		priv, err := ecdsa.GenerateKey(secp256k1.S256(), rand.Reader)
		require.NoError(t, err)

		signer, err := signer.NewECDSASigner(priv)
		require.NoError(t, err)
		assert.NotNil(t, signer)
	})

	t.Run("fails - invalid curve", func(t *testing.T) {
		t.Parallel()

		priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		require.NoError(t, err)

		_, err = signer.NewECDSASigner(priv)
		require.ErrorIs(t, err, signer.ErrInvalidCurve)
	})
}

func TestECDSASignerFromHex(t *testing.T) {
	t.Parallel()

	t.Run("passes- valid hex for secp256k1 private key", func(t *testing.T) {
		t.Parallel()

		privHex, ok := os.LookupEnv("X402_BUYER_PRIVATE_KEY")
		require.True(t, ok)

		signer, err := signer.NewECDSASignerFromHex(privHex)
		require.NoError(t, err)

		apitest.TestSigner(t, signer)
	})

	t.Run("fails - point coordinates not on secp256k1 curve", func(t *testing.T) {
		t.Parallel()

		_, err := signer.NewECDSASignerFromHex(apitest.NotOnCurvePrivateKeyHex)
		require.ErrorIs(t, err, signer.ErrInvalidPoint)
	})
}
