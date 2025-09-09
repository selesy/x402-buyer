# x402-buyer

![x402 buyer](https://pkg.go.dev/badge/github.com/selesy/x402-buyer.svg) ![GitHub Actions Workflow Status](https://img.shields.io/github/actions/workflow/status/selesy/x402-buyer/pre-commit.yaml) ![x402 buyer](https://goreportcard.com/badge/github.com/selesy/x402-buyer) ![readme%20style standard brightgreen](https://img.shields.io/badge/readme%20style-standard-brightgreen.svg?style=flat-square) ![GitHub License](https://img.shields.io/github/license/selesy/x402-buyer) ![GitHub Release](https://img.shields.io/github/v/release/selesy/x402-buyer) ![Conventional Commits](https://img.shields.io/badge/Conventional%20Commits-1.0.0-yellow.svg) ![pre—​commit enabled brightgreen?logo=pre commit](https://img.shields.io/badge/pre—​commit-enabled-brightgreen?logo=pre-commit)

Package `buyer` produces [http.Client](https://pkg.go.dev/net/http#Client)'s that can make [x402 payments](https://x402.org) for HTTP content and services.

## Install

Include this library in your project using the following command:

``` bash
go get github.com/selesy/x402-buyer
```

## Usage

Create an `http.Client` as shown below, then use it to make HTTP requests as usual. If an `x402` payment is required, it will be made by the client and the response will be returned as usual.

``` go
package main

import (
    "io"
    "log/slog"
    "os"

    "github.com/lmittmann/tint"

    buyer "github.com/selesy/x402-buyer"
)

func main() {
    const (
        privateKeyEnvVar = "X402_BUYER_PRIVATE_KEY"
        url              = "https://x402.smoyer.dev/premium-joke"
    )

    log := slog.New(tint.NewHandler(os.Stderr, &tint.Options{
        Level: slog.LevelDebug,
    }))

    client, err := buyer.ClientForPrivateKeyHexFromEnv(privateKeyEnvVar, buyer.WithLogger(log))
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
```

Full documentation for this library is available as [Go docs](https://pkg.go.dev/github.com/selesy/x402-buyer).

## Contributing

- Please report issues using [GitHub Issues](https://github.com/selesy/x402-buyer/issues).

- PRs are happily considered when submitted to [GitHub Pull requests](https://github.com/selesy/x402-buyer/pulls).

- Other questions or discussions can be submitted to [GitHub Discussions](https://github.com/selesy/x402-buyer/discussions).

### Development

This project strives to maintain minimal external dependencies. If you have a feature that requires specific libraries, let’s discuss whether a new Go module should be created in a sub-directory.

The tools required to develop this project and to run the `pre-commit` checks are defined in the `.tool-versions` file.

    asciidoctorj 3.0.0
    golang 1.24.4
    golangci-lint 2.4.0
    pre-commit 4.2.0
    pandoc 3.7.0.2
    python 3.10.4

If you’re using `asdf`, simply run `asdf install`. Otherwise, install the listed tools in the manner required by your operating system. Once the required tools are installed, install the `pre-commit` hooks by running `pre-commit install --install-hooks`. Test your environment by running `pre-commit run --all-files`.

## License

This project is distributed under the [MIT License](https://github.com/selesy/x402-buyer/blob/main/LICENSE).
