package signer_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/selesy/x402-buyer/internal/signer"
	"github.com/selesy/x402-buyer/pkg/api/apitest"
)

func TestKeystoreSigner(t *testing.T) {
	t.Parallel()

	ks, acct := apitest.Keystore(t)

	signer, err := signer.NewKeyStoreSigner(ks, acct, []byte(apitest.Passphrase))
	require.NoError(t, err)

	apitest.TestSigner(t, signer)
}
