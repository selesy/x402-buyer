package evm_test

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gotest.tools/v3/golden"

	"github.com/selesy/x402-buyer/internal/exact/evm"
	"github.com/selesy/x402-buyer/internal/signer"
	"github.com/selesy/x402-buyer/pkg/api/apitest"
	"github.com/selesy/x402-buyer/pkg/payer"
)

func TestNewClient(t *testing.T) {
	t.Parallel()

	paymentRequestJSON := golden.Get(t, "x402_org_payment_request.json")

	signer, err := signer.NewECDSASignerFromHex(apitest.ECDSAPrivateKeyHex)
	require.NoError(t, err)

	var paymentRequest payer.PaymentRequest

	require.NoError(t, json.Unmarshal(paymentRequestJSON, &paymentRequest))
	require.NotNil(t, paymentRequest)
	require.NotNil(t, paymentRequest.Accepts)
	require.Len(t, paymentRequest.Accepts, 1)

	log := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))

	payer, err := evm.NewExactEvm(signer, fixedNowFunc(t), fixedNonceFunc(t), log)
	require.NoError(t, err)

	paymentPayload, err := payer.Pay(paymentRequest.Accepts[0])
	require.NoError(t, err)

	data, err := json.Marshal(paymentPayload)
	require.NoError(t, err)

	buf := &bytes.Buffer{}
	require.NoError(t, json.Indent(buf, data, "", "  "))

	golden.Assert(t, buf.String()+"\n", "x402_org_payment_payload.golden")
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
