package buyer

import "log/slog"

type config struct {
	log *slog.Logger
}

type Option func(*config) error

func WithLogger(log *slog.Logger) Option {
	return func(c *config) error {
		c.log = log

		return nil
	}
}
