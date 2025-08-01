package payer

import (
	"errors"
	"fmt"
)

var ErrFailedPayloadCreate = errors.New("failed to create PaymentPayload")

func FailedPaymentPayloadCreation(err error) error {
	return fmt.Errorf("%w: %w", ErrFailedPayloadCreate, err)
}
