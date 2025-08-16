package apitest

import (
	"crypto/ecdsa"
	"encoding/hex"
	"math/big"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/selesy/x402-buyer/pkg/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	ECDSAPrivateKeyHex      = "6cfb3f917efa513636a6f8103d01426e932806cc7205c4361de4c633452e2b57"
	NotOnCurvePrivateKeyHex = "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"

	Passphrase = "LetMeIn"
)

func TestSigner(t *testing.T, signer api.Signer) {
	const expSig = "1bb6d051bf1a3d8e239b1dc6d92bba5db06ffe2f807e6cdf6555bf4ae801fca6146403a7f9c84bbab9f04fdaf4e46de5784632299cae77d41c71b1038b9e3f5a1b"

	hash, _ := TransferWithAuthorizationHash(t)

	actSig, err := signer.Sign(hash)
	require.NoError(t, err)

	actSig[64] += 27
	assert.Equal(t, expSig, hex.EncodeToString(actSig))
}

func TestDataSigner(t *testing.T, signer api.Signer) {
	const expSig = "1bb6d051bf1a3d8e239b1dc6d92bba5db06ffe2f807e6cdf6555bf4ae801fca6146403a7f9c84bbab9f04fdaf4e46de5784632299cae77d41c71b1038b9e3f5a1b"

	_, data := TransferWithAuthorizationHash(t)

	actSig, err := signer.Sign([]byte(data))
	require.NoError(t, err)

	actSig[64] += 27
	assert.Equal(t, expSig, hex.EncodeToString(actSig))
}

func PrivateKey(t *testing.T) *ecdsa.PrivateKey {
	t.Helper()

	privHex, ok := os.LookupEnv("X402_BUYER_PRIVATE_KEY")
	require.True(t, ok)
	privBytes, err := hex.DecodeString(privHex)
	require.NoError(t, err)

	priv := new(ecdsa.PrivateKey)
	priv.D = new(big.Int).SetBytes(privBytes)
	priv.PublicKey.Curve = secp256k1.S256()
	priv.PublicKey.X, priv.PublicKey.Y = secp256k1.S256().ScalarBaseMult(priv.D.Bytes())
	require.False(t, priv.X == nil || priv.Y == nil)

	return priv
}

func Keystore(t *testing.T) (*keystore.KeyStore, accounts.Account) {
	t.Helper()

	priv := PrivateKey(t)

	path := t.TempDir()
	ks := keystore.NewKeyStore(path, keystore.StandardScryptN, keystore.StandardScryptP)

	acct, err := ks.ImportECDSA(priv, Passphrase)
	require.NoError(t, err)
	require.NotNil(t, acct)

	return ks, acct
}

func Wallet(t *testing.T) (accounts.Wallet, accounts.Account) {
	t.Helper()

	ks, acct := Keystore(t)
	mgr := accounts.NewManager(&accounts.Config{InsecureUnlockAllowed: false}, ks)

	wal, err := mgr.Find(acct)
	require.NoError(t, err)
	require.NotNil(t, wal)

	return wal, acct
}

// func ECDSAPrivateKey(t *testing.T) *ecdsa.PrivateKey {
// 	t.Helper()

// 	return ECDSAPrivateKeyFromHex(t, ECDSAPrivateKeyHex)
// }

// func ECDSAPrivateKeyFromHex(t *testing.T, Hex string) *ecdsa.PrivateKey {
// 	t.Helper()

// 	privBytes, err := hex.DecodeString(Hex)
// 	if err != nil {
// 		panic(err)
// 	}

// 	//     hexString := "your_hex_string_here"
// 	// privateKeyBytes, err := hex.DecodeString(hexString)
// 	// if err != nil {
// 	//     log.Fatal(err)
// 	// }
// 	// privateKey, err := ecdsa.UnmarshalECPrivateKey(privateKeyBytes)
// 	// if err != nil {
// 	//     log.Fatal(err)
// 	// }

// 	privateKey := new(ecdsa.PrivateKey)
// 	privateKey.PublicKey.Curve = crypto.S256()      // Choose the appropriate curve
// 	privateKey.D = new(big.Int).SetBytes(privBytes) // Set the private key value
// 	_ = privateKey

// 	// Generate the public key from the private key
// 	// privateKey.PublicKey.X, privateKey.PublicKey.Y = privateKey.Curve.ScalarBaseMult(privateKey.D.Bytes())
// 	// ecdh.PrivateKeyFromBytes(privBytes)

// 	priv := secp256k1.PrivKeyFromBytes(privBytes).ToECDSA()
// 	_ = priv

// 	assert.Equal(t, priv, privateKey)
// 	// t.Fail()

// 	// pr, err := crypto.HexToECDSA(ECDSAPrivateKeyHex)
// 	pr, err := crypto.ToECDSA(privBytes)
// 	assert.Equal(t, pr, privateKey)

// 	return privateKey
// }

func TransferWithAuthorizationHash(t *testing.T) ([]byte, string) {
	t.Helper()

	const (
		expHash = "291ea3849c8018ce32bbf62d479dc3ddf6aeb48ff26ce781af4c5eaa83279a5a"
		expData = "190102fa7265e7c5d81118673727957699e4d68f74cd74b7db77da710fe8a2c7834f4ef85a66e9f161738930fbdba8ae123e7abd7bd10ce397381f794ad74073192f"
	)

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
			Name:              "USD Coin",
			Version:           "2",
			ChainId:           math.NewHexOrDecimal256(8453),
			VerifyingContract: "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913",
		},
		Message: apitypes.TypedDataMessage{
			"from":        "0x26279EC7Ad9207013149967b5aA1CF42AC6487eb",
			"to":          "0x8d6Efb97F6E3d218647eD74AF418d47489550Ae2",
			"value":       "320",
			"validAfter":  "1754735643",
			"validBefore": "1754736303",
			"nonce":       "0xd8ac8930d08bfa8ff03af000ef78f0c624f30047d52e62b3ae8e3b9e2b6462ca",
		},
	}

	hash, data, err := apitypes.TypedDataAndHash(typedData)
	require.NoError(t, err)
	require.Equal(t, expHash, hex.EncodeToString(hash))
	require.Equal(t, expData, hex.EncodeToString([]byte(data)))

	hash2 := crypto.Keccak256Hash([]byte(data))
	require.Equal(t, expHash, hex.EncodeToString(hash2.Bytes()))

	t.Log("hash:", hex.EncodeToString(hash))
	t.Log("hash2:", hex.EncodeToString(hash2.Bytes()))

	return hash, data
}
