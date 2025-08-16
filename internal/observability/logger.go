package observability

import (
	"context"
	"log/slog"
)

var _ slog.Handler = (*NoopHandler)(nil)

type NoopHandler struct{}

func NewNoopHandler() slog.Handler {
	return &NoopHandler{}
}

func (h *NoopHandler) Enabled(_ context.Context, _ slog.Level) bool {
	return false
}
func (h *NoopHandler) Handle(_ context.Context, _ slog.Record) error {
	return nil
}

func (h *NoopHandler) WithAttrs(_ []slog.Attr) slog.Handler {
	return h
}

func (h *NoopHandler) WithGroup(_ string) slog.Handler {
	return h
}
