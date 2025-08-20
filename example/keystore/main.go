package main

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/lmittmann/tint"

	buyer "github.com/selesy/x402-buyer"
)

func main() {
	const (
		accountAddressEnvVar  = "X402_ACCOUNT_ADDRESS"
		accountPasswordEnvVar = "X402_ACCOUNT_PASSWORD" //nolint:gosec
		keystoreDirectory     = "X402_KEYSTORE_DIRECTORY"
		url                   = "https://x402.smoyer.dev/premium-joke"
	)

	log := slog.New(tint.NewHandler(os.Stderr, &tint.Options{
		Level: slog.LevelDebug,
	}))

	userDir, err := os.UserHomeDir()
	if err != nil {
		log.Error("failed to get user home directory", tint.Err(err))
		os.Exit(1)
	}

	ksPath := filepath.Join(userDir, ".ethereum", "keystore")
	if ks, ok := os.LookupEnv(keystoreDirectory); ok {
		ksPath = ks
	}

	addr, ok := os.LookupEnv(accountAddressEnvVar)
	if !ok {
		log.Error("failed to look up account address environment variable")
		os.Exit(1)
	}

	pass, ok := os.LookupEnv(accountPasswordEnvVar)
	if !ok {
		log.Error("failed to look up account password environment variable")
		os.Exit(1)
	}

	ks := keystore.NewKeyStore(ksPath, keystore.StandardScryptN, keystore.StandardScryptP)
	acct := accounts.Account{Address: common.HexToAddress(addr)}

	client, err := buyer.ClientForKeyStore(ks, acct, []byte(pass), buyer.WithLogger(log))
	if err != nil {
		log.Error("failed to create client", tint.Err(err))
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
