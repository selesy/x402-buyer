package buyer

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/coinbase/x402/go/pkg/types"
	"github.com/lmittmann/tint"

	"github.com/selesy/x402-buyer/internal/exact/evm"
	"github.com/selesy/x402-buyer/pkg/api"
)

var _ http.RoundTripper = (*Transport)(nil)

// Transport is an http.RoundTripper that is capable of making x402 payments
// to access HTTP-based content or services on the Internet.
type Transport struct {
	config

	next   http.RoundTripper
	signer api.Signer
}

// NewTransport creates an http.RoundTripper that is capable of making x402
// payments using the provided api.Signer by wrapping the underlying
// http.Transport provided by the next argument.
func NewTransport(next http.RoundTripper, signer api.Signer, opts ...Option) (*Transport, error) {
	cfg, err := newConfig(opts...)
	if err != nil {
		return nil, err
	}

	return newTransport(next, signer, cfg), nil
}

func newTransport(next http.RoundTripper, signer api.Signer, cfg *config) *Transport {
	return &Transport{
		config: *cfg,

		next:   next,
		signer: signer,
	}
}

// RoundTrip implements http.RoundTripper.
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Body can only be read one time ... since we make two round-trips
	// if a payment is required, we have to duplicate the body.  So we
	// read the bytes and will create new readers for each call.

	var (
		body []byte
		err  error
	)

	if req.Body != nil {
		defer func() {
			if err := req.Body.Close(); err != nil {
				t.log.Error("failed to close request body", tint.Err(err))
			}
		}()

		body, err = io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}

	}

	// Perform the http.Request
	if req.Body != nil {
		req.Body = io.NopCloser(bytes.NewReader(body))
	}

	resp, err := t.next.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	// Return the http.Response if no payment is required
	if resp.StatusCode != http.StatusPaymentRequired {
		return resp, nil
	}

	// Intercept the response with a copy of the request
	if req.Body != nil {
		req.Body = io.NopCloser(bytes.NewReader(body))
	}

	return t.handlePaymentRequired(req, resp)
}

func (t *Transport) handlePaymentRequired(req *http.Request, resp *http.Response) (*http.Response, error) {
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.log.Error("failed to close response body", tint.Err(err))
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	t.log.Debug("Payment request body", slog.String("json", string(body)))

	var paymentRequest api.PaymentRequest
	if err := json.Unmarshal(body, &paymentRequest); err != nil {
		return nil, fmt.Errorf("failed to unmarshal payment request: %w", err)
	}

	if len(paymentRequest.Accepts) == 0 {
		return nil, fmt.Errorf("no payment methods accepted")
	}

	// TODO: For simplicity, we'll just use the first accepted payment method.
	paymentDetails := paymentRequest.Accepts[0]

	payment, err := t.createPayment(paymentDetails)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}

	paymentData, err := json.Marshal(payment)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payment: %w", err)
	}

	t.log.Debug("Payment header JSON", slog.String("json", string(paymentData)))

	req.Header.Set("X-Payment", base64.StdEncoding.EncodeToString(paymentData))

	return t.next.RoundTrip(req)
}

func (t *Transport) createPayment(details types.PaymentRequirements) (*types.PaymentPayload, error) {
	payer, err := evm.NewExactEvm(t.signer, time.Now, api.DefaultNonce, t.log)
	if err != nil {
		return nil, err
	}

	return payer.Pay(details)
}
