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
	"github.com/selesy/x402-buyer/pkg/api/apitest"
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

	cl, err := buyer.ClientForPrivateKeyHex(apitest.ECDSAPrivateKeyHex)
	require.NoError(t, err)
	assert.NotNil(t, cl)

	// TODO: yes we built a client but is it working?
}

func TestClientFromPrivateKeyHexFromEnv(t *testing.T) {
	t.Parallel()

	require.NoError(t, os.Setenv(apitest.ECDSAPrivateKeyHexEnvVarName, apitest.ECDSAPrivateKeyHex))

	cl, err := buyer.ClientForPrivateKeyHexFromEnv(apitest.ECDSAPrivateKeyHexEnvVarName)
	require.NoError(t, err)
	assert.NotNil(t, cl)

	// TODO: yes we built a client but is it working?
}

func TestClientForSigner(t *testing.T) {
	t.Parallel()

	require.NoError(t, os.Setenv(apitest.ECDSAPrivateKeyHexEnvVarName, apitest.ECDSAPrivateKeyHex))
	privHex, ok := os.LookupEnv(apitest.ECDSAPrivateKeyHexEnvVarName)
	require.True(t, ok)

	signer, err := signer.NewECDSASignerFromHex(privHex)
	require.NoError(t, err)

	cl, err := buyer.ClientForSigner(signer)
	require.NoError(t, err)
	assert.NotNil(t, cl)

	// TODO: yes we built a client but is it working?
}
