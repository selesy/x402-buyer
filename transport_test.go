package buyer_test

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	buyer "github.com/selesy/x402-buyer"
	"github.com/selesy/x402-buyer/internal/signer"
	"github.com/selesy/x402-buyer/pkg/api/apitest"
)

func TestTransport(t *testing.T) {
	const payReq = `{"accepts":[{"scheme":"exact","network":"base","maxAmountRequired":"10000","resource":"https://example.com","description":"A premium programming joke","mimeType":"","payTo":"0x60ac86571E55F9735F00cE9e28361d203977B260","maxTimeoutSeconds":60,"asset":"0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913","extra":{"name":"USD Coin","version":"2"}}],"error":"X-PAYMENT header is required","x402Version":1}`

	signer, err := signer.NewECDSASignerFromHex(apitest.ECDSAPrivateKeyHex)
	require.NoError(t, err)

	t.Run("passes - no payment required", func(t *testing.T) {
		t.Parallel()

		respIn1 := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader("Response body")),
		}

		next := newMockTransport(t, respIn1)
		trans, err := buyer.NewTransport(next, signer)
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodGet, "https://example.com", strings.NewReader("Request body"))
		require.NoError(t, err)

		respOut, err := trans.RoundTrip(req)
		require.NoError(t, err)
		assert.Equal(t, respIn1, respOut)

		t.Cleanup(func() {
			require.NoError(t, respOut.Body.Close())
		})
	})

	t.Run("passes - payment required", func(t *testing.T) {
		t.Parallel()

		respIn1 := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader("Response body")),
		}

		respIn2 := &http.Response{
			StatusCode: http.StatusPaymentRequired,
			Body:       io.NopCloser(strings.NewReader(payReq)),
		}

		next := newMockTransport(t, respIn2, respIn1)
		trans, err := buyer.NewTransport(next, signer)
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodGet, "https://example.com", strings.NewReader("Request body"))
		require.NoError(t, err)

		respOut, err := trans.RoundTrip(req)
		require.NoError(t, err)
		assert.Equal(t, respIn1, respOut)

		t.Cleanup(func() {
			require.NoError(t, respOut.Body.Close())
		})
	})

	t.Run("fails - invalid payment header", func(t *testing.T) {
		t.Parallel()

		respIn2 := &http.Response{
			StatusCode: http.StatusPaymentRequired,
			Body:       io.NopCloser(strings.NewReader(payReq)),
		}

		respIn3 := &http.Response{
			StatusCode: http.StatusPaymentRequired,
			Body:       io.NopCloser(strings.NewReader(payReq)),
		}

		next := newMockTransport(t, respIn2, respIn3)
		trans, err := buyer.NewTransport(next, signer)
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodGet, "https://example.com", strings.NewReader("Request body"))
		require.NoError(t, err)

		respOut, err := trans.RoundTrip(req)
		require.NoError(t, err)
		assert.Equal(t, respIn3, respOut)
	})
}

var _ http.RoundTripper = (*mockTransport)(nil)

type mockTransport struct {
	t     *testing.T
	resps []*http.Response
	idx   int
}

func newMockTransport(t *testing.T, resps ...*http.Response) *mockTransport {
	t.Helper()

	return &mockTransport{
		t:     t,
		resps: resps,
		idx:   0,
	}
}

func (t *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	defer func() {
		require.NoError(t.t, req.Body.Close())
	}()

	body, err := io.ReadAll(req.Body)
	require.NoError(t.t, err)
	require.Equal(t.t, "Request body", string(body))

	require.False(t.t, t.idx >= len(t.resps), "Why?")

	out := t.resps[t.idx]
	t.idx++

	return out, nil
}
