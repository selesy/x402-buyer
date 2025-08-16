package buyer

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/selesy/x402-buyer/internal/observability"
)

type config struct {
	client *http.Client
	log    *slog.Logger
}

// Option represents a means of altering the default configuration of the
// buyer's http.RoundTripper.
type Option func(*config) error

func newConfig(opts ...Option) (*config, error) {
	var errs error

	cfg := &config{
		client: &http.Client{
			Transport: http.DefaultTransport,
		},
		log: slog.New(observability.NewNoopHandler()),
	}

	for _, opt := range opts {
		errs = errors.Join(errs, opt(cfg))
	}

	if errs != nil {
		return nil, errs
	}

	return cfg, nil
}

// WithClient is an Option that allows the user to provide a custom http.Client
// whose http.RoundTripper will be wrapped to allow x402 payments.
//
// If not provided, http.DefaultClient will be used and internally, the
// http.DefaultTransport will be wrapped with the payment middleware.  This
// option is ignored when provided as an argument to NewTransport.
func WithClient(client *http.Client) Option {
	return func(c *config) error {
		c.client = client

		return nil
	}
}

// WithLogger is an Option that allows the user to provide an slog.Logger that
// can be used to observe the internal operation of the buyer's http.RoundTripper.
//
// If not provided, a No-Op logger is used.  Under normal operation, this library
// writes one line of INFO-level logging for each payment that's made.  Debug-
// level logging provides a log record for each step in the payment process.
func WithLogger(log *slog.Logger) Option {
	return func(c *config) error {
		c.log = log

		return nil
	}
}
