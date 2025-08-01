package buyer

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/coinbase/x402/go/pkg/types"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/selesy/x402-buyer/internal/exact/evm"
	"github.com/selesy/x402-buyer/pkg/payer"
)

var _ http.RoundTripper = (*X402BuyerTransport)(nil)

type X402BuyerTransport struct {
	config

	next http.RoundTripper
	priv *ecdsa.PrivateKey
	wal  accounts.Wallet
	acct accounts.Account
}

func NewX402BuyerTransport(next http.RoundTripper, priv *ecdsa.PrivateKey, wal accounts.Wallet, acct accounts.Account, opts ...Option) (*X402BuyerTransport, error) {
	var errs error

	cfg := config{
		log: slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{})),
	}

	for _, opt := range opts {
		if err := opt(&cfg); err != nil {
			errs = errors.Join(errs, err)
		}
	}

	if errs != nil {
		return nil, errs
	}

	return &X402BuyerTransport{
		config: cfg,

		next: next,
		priv: priv,
		wal:  wal,
		acct: acct,
	}, nil
}

func (t *X402BuyerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Body can only be read one time ... since we make two round-trips
	// if a payment is required, we have to duplicate the body.  So we
	// read the bytes and will create new readers for each call.

	var (
		body []byte
		err  error
	)

	if req.Body != nil {
		defer req.Body.Close()

		body, err = io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}

	}

	if req.Body != nil {
		req.Body = io.NopCloser(bytes.NewReader(body))
	}

	resp, err := t.next.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusPaymentRequired {
		return resp, nil
	}

	if req.Body != nil {
		req.Body = io.NopCloser(bytes.NewReader(body))
	}

	return t.handlePaymentRequired(req, resp)
}

func (t *X402BuyerTransport) handlePaymentRequired(req *http.Request, resp *http.Response) (*http.Response, error) {
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	t.log.Debug("Payment request body", slog.String("json", string(body)))

	var paymentRequest payer.PaymentRequest
	if err := json.Unmarshal(body, &paymentRequest); err != nil {
		return nil, fmt.Errorf("failed to unmarshal payment request: %w", err)
	}

	if len(paymentRequest.Accepts) == 0 {
		return nil, fmt.Errorf("no payment methods accepted")
	}

	// TODO: For simplicity, we'll just use the first accepted payment method.
	paymentDetails := paymentRequest.Accepts[0]

	payment, err := t.createPayment(paymentDetails)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}

	paymentData, err := json.Marshal(payment)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payment: %w", err)
	}

	t.log.Debug("Payment header JSON", slog.String("json", string(paymentData)))

	req.Header.Set("X-Payment", base64.StdEncoding.EncodeToString(paymentData))

	return t.next.RoundTrip(req)
}

func (t *X402BuyerTransport) createPayment(details types.PaymentRequirements) (*types.PaymentPayload, error) {
	payer, err := evm.NewExactEvm(t.priv, time.Now, payer.DefaultNonce, t.log)
	if err != nil {
		return nil, err
	}

	return payer.Pay(details)
	// switch details.Scheme {
	// case "exact":
	// 	return t.createPaymentExactEvm(details)
	// default:
	// 	return nil, fmt.Errorf("unknown payment scheme : %s: %w", http.ErrNotSupported, details.Scheme)
	// }
}

// func (t *X402BuyerTransport) createPaymentExactEvm(details types.PaymentRequirements) (*types.PaymentPayload, error) {
// 	payload, err := t.preparePaymentHeader(details)
// 	if err != nil {
// 		return nil, err
// 	}

// 	var extra map[string]any

// 	if err := json.Unmarshal([]byte(*details.Extra), &extra); err != nil {
// 		return nil, err
// 	}

// 	chain, ok := map[string]*math.HexOrDecimal256{
// 		"base":         math.NewHexOrDecimal256(8453),
// 		"base-sepolia": math.NewHexOrDecimal256(84532),
// 	}[details.Network]

// 	if !ok {
// 		return nil, fmt.Errorf("unknown network: %s", details.Network)
// 	}

// 	typedData := apitypes.TypedData{
// 		Types: apitypes.Types{
// 			"TransferWithAuthorization": []apitypes.Type{
// 				{Name: "from", Type: "address"},
// 				{Name: "to", Type: "address"},
// 				{Name: "value", Type: "uint256"},
// 				{Name: "validAfter", Type: "uint256"},
// 				{Name: "validBefore", Type: "uint256"},
// 				{Name: "nonce", Type: "bytes32"},
// 			},
// 			"EIP712Domain": []apitypes.Type{
// 				{Name: "name", Type: "string"},
// 				{Name: "version", Type: "string"},
// 				{Name: "chainId", Type: "uint256"},
// 				{Name: "verifyingContract", Type: "address"},
// 			},
// 		},
// 		PrimaryType: "TransferWithAuthorization",
// 		Domain: apitypes.TypedDataDomain{
// 			Name: extra["name"].(string), // TODO
// 			// Name: "USDC",
// 			Version: extra["version"].(string), // TODO
// 			// Version: "2",
// 			// ChainId: math.NewHexOrDecimal256(8453), // mainnet
// 			// ChainId:           math.NewHexOrDecimal256(84532), // testnet
// 			ChainId:           chain,
// 			VerifyingContract: details.Asset,
// 		},
// 		Message: apitypes.TypedDataMessage{
// 			"from":        payload.Payload.Authorization.From,
// 			"to":          payload.Payload.Authorization.To,
// 			"value":       payload.Payload.Authorization.Value,
// 			"validAfter":  payload.Payload.Authorization.ValidAfter,
// 			"validBefore": payload.Payload.Authorization.ValidBefore,
// 			"nonce":       payload.Payload.Authorization.Nonce,
// 		},
// 	}

// 	// typedDataJSON, err := json.Marshal(typedData)
// 	// if err != nil {
// 	// 	return nil, err
// 	// }

// 	// fmt.Println("Encoded:", typedData.EncodeType("TransferWithAuthorization"))

// 	// domainSeparator, err := typedData.HashStruct("EIP712Domain", typedData.Domain.Map())
// 	// if err != nil {
// 	// 	return nil, err
// 	// }

// 	// typedDataHash, err := typedData.HashStruct(typedData.PrimaryType, typedData.Message)
// 	// if err != nil {
// 	// 	log.Fatal(err)
// 	// }

// 	// signHash := crypto.Keccak256Hash(
// 	// 	[]byte{0x19, 0x01},
// 	// 	domainSeparator,
// 	// 	typedDataHash,
// 	// )

// 	jsonData, err := json.Marshal(typedData)
// 	if err != nil {
// 		return nil, err
// 	}

// 	fmt.Println("JSON:", string(jsonData))
// 	fmt.Println("JSON hex:", hexutil.Encode(jsonData))

// 	hash, data, err := apitypes.TypedDataAndHash(typedData)
// 	if err != nil {
// 		return nil, err
// 	}

// 	fmt.Println("Hash:", hexutil.Encode(hash))
// 	fmt.Println("Message:", data)
// 	fmt.Println("Message hex:", hexutil.Encode([]byte(data)))

// 	// sig, err := t.ks.SignHash(t.acct, hash)
// 	// if err != nil {
// 	// 	return nil, err
// 	// }

// 	sig1, err := t.wal.SignData(t.acct, "", []byte(data))
// 	if err != nil {
// 		return nil, err
// 	}

// 	sig1[64] += 27

// 	// fmt.Println("Signature:", hexutil.Encode(sig))
// 	// fmt.Println("Signature:", hexutil.Encode(sig1))
// 	fmt.Println("Signature:", hex.EncodeToString(sig1))

// 	// payload.Payload.Signature = hex.EncodeToString(sig1)
// 	// payload.Payload.Signature = common.Bytes2Hex(sig)
// 	payload.Payload.Signature = hexutil.Encode(sig1)

// 	sig1[64] -= 27

// 	// pubKey, err := crypto.Ecrecover(hash, sig)
// 	pubKey, err := crypto.SigToPub(hash, sig1)
// 	if err != nil {
// 		return nil, err
// 	}

// 	addr := crypto.PubkeyToAddress(*pubKey)
// 	// recoveredAddress := crypto.PubkeyToAddress(*crypto.ToECDSAPub(pubKey))

// 	// fmt.Println("Recoverred public key:", hexutil.Encode(pubKey.))
// 	fmt.Println("Recovered Address:", addr.Hex())

// 	fmt.Println("Payload:", payload)
// 	fmt.Println("Exact EVM:", payload.Payload)
// 	fmt.Println("Exact EVM Authorization:", payload.Payload.Authorization)

// 	return payload, nil
// }

// func (t *X402BuyerTransport) preparePaymentHeader(details types.PaymentRequirements) (*types.PaymentPayload, error) {
// 	nonce := make([]byte, 32)
// 	if _, err := rand.Read(nonce); err != nil {
// 		return nil, err
// 	}

// 	validAfter := strconv.FormatInt(time.Now().Add(-10*time.Minute).Unix(), 10)
// 	validBefore := strconv.FormatInt(time.Now().Add(time.Duration(details.MaxTimeoutSeconds)*time.Second).Unix(), 10)

// 	return &types.PaymentPayload{
// 		X402Version: 1,
// 		Scheme:      details.Scheme,
// 		Network:     details.Network,
// 		Payload: &types.ExactEvmPayload{
// 			Signature: "",
// 			Authorization: &types.ExactEvmPayloadAuthorization{
// 				From:        t.acct.Address.String(),
// 				To:          details.PayTo,
// 				Value:       details.MaxAmountRequired,
// 				ValidAfter:  validAfter,
// 				ValidBefore: validBefore,
// 				Nonce:       hexutil.Encode(nonce),
// 			},
// 		},
// 	}, nil
// }
