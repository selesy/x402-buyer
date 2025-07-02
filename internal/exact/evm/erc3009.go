package evm

import (
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/coinbase/x402/go/pkg/types"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/selesy/x402-buyer/pkg/payer"
)

var _ payer.Payer = (*ExactEvm)(nil)

type ExactEvm struct {
	wal       accounts.Wallet
	acct      accounts.Account
	priv      *ecdsa.PrivateKey
	nowFunc   payer.NowFunc
	nonceFunc payer.NonceFunc
}

func NewExactEvm(priv *ecdsa.PrivateKey, nowFunc payer.NowFunc, nonceFunc payer.NonceFunc) (*ExactEvm, error) {
	path := filepath.Join(os.TempDir(), "x402", "buyer", "keystore")

	if err := os.MkdirAll(path, 0o700); err != nil {
		return nil, err
	}

	ks := keystore.NewKeyStore(path, keystore.StandardScryptN, keystore.StandardScryptP)

	_, err := ks.ImportECDSA(priv, "")
	if err != nil && !errors.Is(err, keystore.ErrAccountAlreadyExists) {
		return nil, err
	}

	if err != nil {
		fmt.Println("Warning:", err.Error())
	}

	pub, _ := priv.Public().(*ecdsa.PublicKey)
	addr := crypto.PubkeyToAddress(*pub)
	acct := accounts.Account{
		Address: addr,
		URL:     accounts.URL{},
	}

	fmt.Println("Original address:", addr.Hex())

	mgr := accounts.NewManager(&accounts.Config{InsecureUnlockAllowed: false}, ks)

	wal, err := mgr.Find(acct)
	if err != nil {
		return nil, err
	}

	if err := ks.Unlock(acct, ""); err != nil {
		return nil, err
	}

	if err := os.RemoveAll(path); err != nil {
		return nil, err
	}

	return &ExactEvm{
		wal:       wal,
		acct:      acct,
		priv:      priv,
		nowFunc:   nowFunc,
		nonceFunc: nonceFunc,
	}, nil
}

// func (e *ExactEvm) Pay(requirements types.PaymentRequirements) (*types.PaymentPayload, payer.Signature, error) {
// 	return nil, "", nil // TODO:
// }

func (e *ExactEvm) Pay(details types.PaymentRequirements) (*types.PaymentPayload, error) {
	switch details.Scheme {
	case "exact":
		return e.createPaymentExactEvm(details)
	default:
		return nil, fmt.Errorf("unknown payment scheme : %w, %s", http.ErrNotSupported, details.Scheme)
	}
}

func (e *ExactEvm) createPaymentExactEvm(details types.PaymentRequirements) (*types.PaymentPayload, error) {
	payload, err := e.preparePaymentHeader(details)
	if err != nil {
		return nil, err
	}

	var extra map[string]any

	if err := json.Unmarshal([]byte(*details.Extra), &extra); err != nil {
		return nil, err
	}

	chain, ok := map[string]*math.HexOrDecimal256{
		"base":         math.NewHexOrDecimal256(8453),
		"base-sepolia": math.NewHexOrDecimal256(84532),
	}[details.Network]

	if !ok {
		return nil, fmt.Errorf("unknown network: %s", details.Network)
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
			},
		},
		PrimaryType: "TransferWithAuthorization",
		Domain: apitypes.TypedDataDomain{
			Name: extra["name"].(string), // TODO
			// Name: "USDC",
			Version: extra["version"].(string), // TODO
			// Version: "2",
			// ChainId:           math.NewHexOrDecimal256(8453), // mainnet
			ChainId:           chain,
			VerifyingContract: details.Asset,
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

	// typedDataJSON, err := json.Marshal(typedData)
	// if err != nil {
	// 	return nil, err
	// }

	// fmt.Println("Encoded:", typedData.EncodeType("TransferWithAuthorization"))

	// domainSeparator, err := typedData.HashStruct("EIP712Domain", typedData.Domain.Map())
	// if err != nil {
	// 	return nil, err
	// }

	// typedDataHash, err := typedData.HashStruct(typedData.PrimaryType, typedData.Message)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// signHash := crypto.Keccak256Hash(
	// 	[]byte{0x19, 0x01},
	// 	domainSeparator,
	// 	typedDataHash,
	// )

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

	// sig, err := t.ks.SignHash(t.acct, hash)
	// if err != nil {
	// 	return nil, err
	// }

	sig1, err := e.wal.SignData(e.acct, "", []byte(data))
	if err != nil {
		return nil, err
	}

	sig1[64] += 27

	// fmt.Println("Signature:", hexutil.Encode(sig))
	// fmt.Println("Signature:", hexutil.Encode(sig1))
	fmt.Println("Signature:", hex.EncodeToString(sig1))

	sig2, err := e.priv.Sign(rand.Reader, hash, nil)
	if err != nil {
		return nil, err
	}

	sig2[64] += 27

	fmt.Println("Signature2:", hex.EncodeToString(sig2))

	sig3, err := crypto.Sign(hash, e.priv)
	if err != nil {
		return nil, err
	}

	sig3[64] += 27

	fmt.Println("Signature3:", hex.EncodeToString(sig3))

	// payload.Payload.Signature = hex.EncodeToString(sig1)
	// payload.Payload.Signature = common.Bytes2Hex(sig)
	payload.Payload.Signature = hexutil.Encode(sig1)

	sig1[64] -= 27
	sig3[64] -= 27

	// pubKey, err := crypto.Ecrecover(hash, sig)
	pubKey, err := crypto.SigToPub(hash, sig3)
	if err != nil {
		return nil, err
	}

	addr := crypto.PubkeyToAddress(*pubKey)
	// recoveredAddress := crypto.PubkeyToAddress(*crypto.ToECDSAPub(pubKey))

	// fmt.Println("Recoverred public key:", hexutil.Encode(pubKey.))
	fmt.Println("Recovered Address:", addr.Hex())

	fmt.Println("Payload:", payload)
	fmt.Println("Exact EVM:", payload.Payload)
	fmt.Println("Exact EVM Authorization:", payload.Payload.Authorization)

	return payload, nil
}

func (e *ExactEvm) preparePaymentHeader(details types.PaymentRequirements) (*types.PaymentPayload, error) {
	// nonce := make([]byte, 32)
	// if _, err := rand.Read(nonce); err != nil {
	// 	return nil, err
	// }

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
				From:        e.acct.Address.String(),
				To:          details.PayTo,
				Value:       details.MaxAmountRequired,
				ValidAfter:  validAfter,
				ValidBefore: validBefore,
				Nonce:       hexutil.Encode(nonce),
			},
		},
	}, nil
}
