package api

import (
	"crypto/rand"
	"time"

	"github.com/coinbase/x402/go/pkg/types"
)

type Scheme string

const (
	SchemeExact Scheme = "exact"
)

// Payer represents types that can be registered and make payments on the
// client's behalf.
type Payer interface {
	// Pay creates a signed types.PaymentPayload for the given
	// types.PaymentRequirements using the private key and configuration
	// provided through it's constructor.
	Pay(requirements types.PaymentRequirements) (*types.PaymentPayload, error)
	// Scheme returns a constant Scheme that the http.RoundTripper can to
	// "route" a payment request to a payer.Payer that can make the
	// appropriate payment.
	Scheme() Scheme
}

type Signature string

// PaymentRequest represents the body of a 402 Payment Required response.
type PaymentRequest struct {
	X402Version int                         `json:"x402Version"`
	Err         string                      `json:"error"`
	Accepts     []types.PaymentRequirements `json:"accepts"`
}

type NonceFunc func() []byte

type NowFunc func() time.Time

func DefaultNonce() []byte {
	nonce := make([]byte, 32)
	_, _ = rand.Read(nonce)

	return nonce
}

func DefaultNow() NowFunc {
	return time.Now
}
