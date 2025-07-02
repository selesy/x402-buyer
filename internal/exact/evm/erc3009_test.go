package evm_test

import (
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/decred/dcrd/dcrec/secp256k1"
	"github.com/selesy/x402-buyer/internal/exact/evm"
	"github.com/selesy/x402-buyer/pkg/payer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// const x402OrgPaymentRequest = `{"x402Version":1,"error":"X-PAYMENT header is required","accepts":[{"scheme":"exact","network":"base-sepolia","maxAmountRequired":"10000","resource":"https://www.x402.org/protected","description":"Access to protected content","mimeType":"application/json","payTo":"0x209693Bc6afc0C5328bA36FaF03C514EF312287C","maxTimeoutSeconds":300,"asset":"0x036CbD53842c5426634e7929541eC2318f3dCF7e","extra":{"name":"USDC","version":"2"}}]}`

func TestNewClient(t *testing.T) {
	t.Parallel()

	paymentRequestJSON, err := os.ReadFile(filepath.Join("testdata", "x402_org_payment_request.json"))
	require.NoError(t, err)

	const exp = "0x6654c039206c904e5a7372069615184ff405c8cc1d50f6e055b69e068b6ad4c33f5572cb913c89a71af37c49c37a391d972032a198b109c41cbf4cbf1473e3fc1c"

	privHex, ok := os.LookupEnv("X402_BUYER_PRIVATE_KEY")
	require.True(t, ok)

	privBytes, err := hex.DecodeString(privHex)
	require.NoError(t, err)

	priv, _ := secp256k1.PrivKeyFromBytes(privBytes)

	var paymentRequest payer.PaymentRequest

	require.NoError(t, json.Unmarshal(paymentRequestJSON, &paymentRequest))
	require.NotNil(t, paymentRequest)
	require.NotNil(t, paymentRequest.Accepts)
	require.Len(t, paymentRequest.Accepts, 1)

	payer, err := evm.NewExactEvm(priv.ToECDSA(), fixedNowFunc(t), fixedNonceFunc(t))
	require.NoError(t, err)

	paymentPayload, err := payer.Pay(paymentRequest.Accepts[0])
	require.NoError(t, err)

	_ = paymentPayload

	assert.Equal(t, exp, string(paymentPayload.Payload.Signature))

	t.Fail()
}

func fixedNonceFunc(t *testing.T) payer.NonceFunc {
	t.Helper()

	nonce, err := hex.DecodeString("140fd607c52d266941aa8d8241891654b6d7ab50a02028cb900c746e3a1bf4dd")
	require.NoError(t, err)

	return func() []byte {
		return nonce
	}
}

func fixedNowFunc(t *testing.T) payer.NowFunc {
	t.Helper()

	now, err := time.Parse(time.RFC3339, "2001-02-03T04:05:06Z")
	require.NoError(t, err)

	return func() time.Time {
		return now
	}
}
