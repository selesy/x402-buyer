package buyer

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/coinbase/x402/go/pkg/types"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
)

var _ http.RoundTripper = (*X402BuyerTransport)(nil)

type X402BuyerTransport struct {
	next http.RoundTripper
	wal  accounts.Wallet
	acct accounts.Account
}

func NewX402BuyerTransport(next http.RoundTripper, wal accounts.Wallet, acct accounts.Account) *X402BuyerTransport {
	return &X402BuyerTransport{
		next: next,
		wal:  wal,
		acct: acct,
	}
}

func (t *X402BuyerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := t.next.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusPaymentRequired {
		return resp, nil
	}

	return t.handlePaymentRequired(req, resp)
}

func (t *X402BuyerTransport) handlePaymentRequired(req *http.Request, resp *http.Response) (*http.Response, error) {
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var paymentRequest PaymentRequest
	if err := json.Unmarshal(body, &paymentRequest); err != nil {
		return nil, fmt.Errorf("failed to unmarshal payment request: %w", err)
	}

	if len(paymentRequest.Accepts) == 0 {
		return nil, fmt.Errorf("no payment methods accepted")
	}

	// For simplicity, we'll just use the first accepted payment method.
	paymentDetails := paymentRequest.Accepts[0]

	payment, err := t.createPayment(paymentDetails)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}

	paymentData, err := json.Marshal(payment)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payment: %w", err)
	}

	// Create a new request with the payment header.
	newReq := req.Clone(req.Context())
	newReq.Header.Set("X-PAYMENT", base64.StdEncoding.EncodeToString(paymentData))
	newReq.Body = io.NopCloser(bytes.NewReader(paymentData))

	return t.next.RoundTrip(newReq)
}

// PaymentRequest represents the body of a 402 Payment Required response.
type PaymentRequest struct {
	Accepts []types.PaymentRequirements `json:"accepts"`
}

func createNonce() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return common.Bytes2Hex(buf), nil
}

func (t *X402BuyerTransport) createPayment(details types.PaymentRequirements) (*types.PaymentPayload, error) {
	return nil, nil
}
