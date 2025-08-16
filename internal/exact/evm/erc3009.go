package evm

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/coinbase/x402/go/pkg/types"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/selesy/x402-buyer/pkg/api"
	"github.com/selesy/x402-buyer/pkg/payer"
)

var _ payer.Payer = (*ExactEvm)(nil)

// ExactEvm is a payer.Payer that handles payment requests on EVM-compatible
// networks for the "exact" scheme.
type ExactEvm struct {
	signer    api.EVMSigner
	nowFunc   payer.NowFunc
	nonceFunc payer.NonceFunc
	log       *slog.Logger
}

func NewExactEvm(signer api.Signer, nowFunc payer.NowFunc, nonceFunc payer.NonceFunc, log *slog.Logger) (*ExactEvm, error) {
	s, ok := signer.(api.EVMSigner)
	if !ok {
		return nil, errors.New("Exact EVM requires an EVM signer")
	}

	return &ExactEvm{
		signer:    s,
		nowFunc:   nowFunc,
		nonceFunc: nonceFunc,
		log:       log,
	}, nil
}

// Pay implements payer.Pay.
func (e *ExactEvm) Pay(requirements types.PaymentRequirements) (*types.PaymentPayload, error) {
	switch requirements.Scheme {
	case "exact":
		return e.createPaymentExactEvm(requirements)
	default:
		return nil, fmt.Errorf("unknown payment scheme : %w, %s", http.ErrNotSupported, requirements.Scheme)
	}
}

// Scheme implements payer.Pay.
func (e *ExactEvm) Scheme() payer.Scheme {
	return payer.SchemeExact
}

func (e *ExactEvm) createPaymentExactEvm(requirements types.PaymentRequirements) (*types.PaymentPayload, error) {
	payload, err := e.preparePaymentHeader(requirements)
	if err != nil {
		return nil, err
	}

	var extra map[string]any

	if err := json.Unmarshal([]byte(*requirements.Extra), &extra); err != nil {
		return nil, err
	}

	chain, ok := map[string]*math.HexOrDecimal256{
		"base":         math.NewHexOrDecimal256(8453),
		"base-sepolia": math.NewHexOrDecimal256(84532),
	}[requirements.Network]

	if !ok {
		return nil, fmt.Errorf("unknown network: %s", requirements.Network)
	}

	typedData := apitypes.TypedData{
		Types: apitypes.Types{
			"TransferWithAuthorization": []apitypes.Type{
				{Name: "from", Type: "address"},
				{Name: "to", Type: "address"},
				{Name: "value", Type: "uint256"},
				{Name: "validAfter", Type: "uint256"},
				{Name: "validBefore", Type: "uint256"},
				{Name: "nonce", Type: "bytes32"},
			},
			"EIP712Domain": []apitypes.Type{
				{Name: "name", Type: "string"},
				{Name: "version", Type: "string"},
				{Name: "chainId", Type: "uint256"},
				{Name: "verifyingContract", Type: "address"},
				// {Name: "salt", Type: "bytes"}, TODO ?
			},
		},
		PrimaryType: "TransferWithAuthorization",
		Domain: apitypes.TypedDataDomain{
			Name:              extra["name"].(string),    // TODO
			Version:           extra["version"].(string), // TODO
			ChainId:           chain,
			VerifyingContract: requirements.Asset,
			// Salt:              "0x", TODO ?
		},
		Message: apitypes.TypedDataMessage{
			"from":        payload.Payload.Authorization.From,
			"to":          payload.Payload.Authorization.To,
			"value":       payload.Payload.Authorization.Value,
			"validAfter":  payload.Payload.Authorization.ValidAfter,
			"validBefore": payload.Payload.Authorization.ValidBefore,
			"nonce":       payload.Payload.Authorization.Nonce,
		},
	}

	hash, data, err := apitypes.TypedDataAndHash(typedData)
	if err != nil {
		return nil, err
	}

	e.log.Debug("ERC-3009 hash", slog.String("hex", hexutil.Encode(hash)))
	e.log.Debug("ERC-3009 message", slog.String("hex", hexutil.Encode([]byte(data))))

	sig, err := e.signer.Sign(hash)
	if err != nil {
		return nil, err
	}

	sig[64] += 27

	e.log.Debug("Signature", slog.String("hex", hex.EncodeToString(sig)))

	payload.Payload.Signature = hexutil.Encode(sig)

	sig[64] -= 27

	pubKey, err := crypto.SigToPub(hash, sig)
	if err != nil {
		return nil, err
	}

	addr := crypto.PubkeyToAddress(*pubKey)

	e.log.Debug("Recovered address", slog.String("hex", addr.Hex()))

	e.log.Info(
		"x402 payment authorized",
		slog.String("from", payload.Payload.Authorization.From),
		slog.String("to", payload.Payload.Authorization.To),
		slog.String("value", payload.Payload.Authorization.Value),
		slog.String("scheme", requirements.Scheme),
		slog.String("network", requirements.Network),
		slog.String("name", extra["name"].(string)),
	)

	return payload, nil
}

func (e *ExactEvm) preparePaymentHeader(details types.PaymentRequirements) (*types.PaymentPayload, error) {
	nonce := e.nonceFunc()

	validAfter := strconv.FormatInt(e.nowFunc().Add(-10*time.Minute).Unix(), 10)
	validBefore := strconv.FormatInt(e.nowFunc().Add(time.Duration(details.MaxTimeoutSeconds)*time.Second).Unix(), 10)

	return &types.PaymentPayload{
		X402Version: 1,
		Scheme:      details.Scheme,
		Network:     details.Network,
		Payload: &types.ExactEvmPayload{
			Signature: "",
			Authorization: &types.ExactEvmPayloadAuthorization{
				From:        e.signer.Address().Hex(),
				To:          details.PayTo,
				Value:       details.MaxAmountRequired,
				ValidAfter:  validAfter,
				ValidBefore: validBefore,
				Nonce:       hexutil.Encode(nonce),
			},
		},
	}, nil
}
