package buyer_test

import (
	"encoding/hex"
	"os"
	"testing"

	"github.com/decred/dcrd/dcrec/secp256k1"
	buyer "github.com/selesy/x402-buyer"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	t.Parallel()

	privHex, ok := os.LookupEnv("X402_BUYER_PRIVATE_KEY")
	require.True(t, ok)

	privBytes, err := hex.DecodeString(privHex)
	require.NoError(t, err)

	priv, _ := secp256k1.PrivKeyFromBytes(privBytes)

	cl, err := buyer.ClientForKey(priv.ToECDSA())
	require.NoError(t, err)

	_ = cl

	t.Fail()
}
