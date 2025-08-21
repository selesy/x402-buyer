package main

import (
	"io"
	"log/slog"
	"os"

	"github.com/lmittmann/tint"

	buyer "github.com/selesy/x402-buyer"
	"github.com/selesy/x402-buyer/third-party/decred"
)

func main() {
	const (
		privateKeyEnvVar = "X402_BUYER_PRIVATE_KEY"
		url              = "https://x402.smoyer.dev/premium-joke"
	)

	log := slog.New(tint.NewHandler(os.Stderr, &tint.Options{
		Level: slog.LevelDebug,
	}))

	client, err := decred.ClientForPrivateKeyHexFromEnv(privateKeyEnvVar, buyer.WithLogger(log))
	if err != nil {
		slog.Error("failed to create client", tint.Err(err))
		os.Exit(1)
	}

	resp, err := client.Get(url)
	if err != nil {
		log.Error("failed to make HTTP request", tint.Err(err))
		os.Exit(1)
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Error("failed to close response body", tint.Err(err))
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	log.Info("HTTP response", slog.String("body", string(body)), slog.Int("code", resp.StatusCode))

	for k, vs := range resp.Header {
		for _, v := range vs {
			log.Debug("HTTP response header", slog.String("key", k), slog.String("value", v))
		}
	}
}
