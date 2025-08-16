package buyer_test

import (
	"crypto/ecdsa"
	"crypto/rand"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	buyer "github.com/selesy/x402-buyer"
	"github.com/selesy/x402-buyer/internal/signer"
)

const (
	testEnvVarName = "X402_BUYER_PRIVATE_KEY"
	testPrivKey    = "7dad518a602e2b504e228012553cc2648109202ebb09f347646e9013b88f22d5"
)

func TestClientForPrivateKey(t *testing.T) {
	t.Parallel()

	priv, err := ecdsa.GenerateKey(secp256k1.S256(), rand.Reader)
	require.NoError(t, err)

	cl, err := buyer.ClientForPrivateKey(priv)
	require.NoError(t, err)
	assert.NotNil(t, cl)

	// TODO: yes we built a client but is it working?
}

func TestClientFromPrivateKeyHex(t *testing.T) {
	t.Parallel()

	cl, err := buyer.ClientForPrivateKeyHex(testPrivKey)
	require.NoError(t, err)
	assert.NotNil(t, cl)

	// TODO: yes we built a client but is it working?
}

func TestClientFromPrivateKeyHexFromEnv(t *testing.T) {
	t.Parallel()

	require.NoError(t, os.Setenv("X402_BUYER_PRIVATE_KEY", testPrivKey))

	cl, err := buyer.ClientForPrivateKeyHexFromEnv(testEnvVarName)
	require.NoError(t, err)
	assert.NotNil(t, cl)

	// TODO: yes we built a client but is it working?
}

func TestClientForSigner(t *testing.T) {
	t.Parallel()

	privHex, ok := os.LookupEnv("X402_BUYER_PRIVATE_KEY")
	require.True(t, ok)

	signer, err := signer.NewECDSASignerFromHex(privHex)
	require.NoError(t, err)

	cl, err := buyer.ClientForSigner(signer)
	require.NoError(t, err)
	assert.NotNil(t, cl)

	// TODO: yes we built a client but is it working?
}
