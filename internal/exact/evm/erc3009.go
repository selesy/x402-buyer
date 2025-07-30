package evm

import (
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/coinbase/x402/go/pkg/types"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/selesy/x402-buyer/pkg/payer"
)

var _ payer.Payer = (*ExactEvm)(nil)

// ExactEvm is a payer.Payer that handles payment requests on EVM-compatible
// networks for the "exact" scheme.
type ExactEvm struct {
	priv      *ecdsa.PrivateKey
	nowFunc   payer.NowFunc
	nonceFunc payer.NonceFunc
}

func NewExactEvm(priv *ecdsa.PrivateKey, nowFunc payer.NowFunc, nonceFunc payer.NonceFunc) (*ExactEvm, error) {
	return &ExactEvm{
		priv:      priv,
		nowFunc:   nowFunc,
		nonceFunc: nonceFunc,
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

	jsonData, err := json.Marshal(typedData)
	if err != nil {
		return nil, err
	}

	fmt.Println("JSON:", string(jsonData))
	fmt.Println("JSON hex:", hexutil.Encode(jsonData))

	hash, data, err := apitypes.TypedDataAndHash(typedData)
	if err != nil {
		return nil, err
	}

	fmt.Println("Hash:", hexutil.Encode(hash))
	fmt.Println("Message:", data)
	fmt.Println("Message hex:", hexutil.Encode([]byte(data)))

	sig3, err := crypto.Sign(hash, e.priv)
	if err != nil {
		return nil, err
	}

	sig3[64] += 27

	fmt.Println("Signature3:", hex.EncodeToString(sig3))

	payload.Payload.Signature = hexutil.Encode(sig3)

	sig3[64] -= 27

	pubKey, err := crypto.SigToPub(hash, sig3)
	if err != nil {
		return nil, err
	}

	addr := crypto.PubkeyToAddress(*pubKey)

	fmt.Println("Recovered Address:", addr.Hex())

	fmt.Println("Payload:", payload)
	fmt.Println("Exact EVM:", payload.Payload)
	fmt.Println("Exact EVM Authorization:", payload.Payload.Authorization)

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
				From:        crypto.PubkeyToAddress(e.priv.PublicKey).Hex(),
				To:          details.PayTo,
				Value:       details.MaxAmountRequired,
				ValidAfter:  validAfter,
				ValidBefore: validBefore,
				Nonce:       hexutil.Encode(nonce),
			},
		},
	}, nil
}
