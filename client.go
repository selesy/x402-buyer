package buyer

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/crypto"
)

func ClientForKey(priv *ecdsa.PrivateKey) (*http.Client, error) {
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

	return ClientForWallet(wal, acct)
}

func ClientForWallet(wal accounts.Wallet, acct accounts.Account) (*http.Client, error) {
	if !wal.Contains(acct) {
		return nil, errors.New("wallet does not contain target account")
	}

	// var err error

	// if err = wallet.Open(""); err != nil {
	// 	return nil, err
	// }

	// defer func() {
	// 	if closeErr := wallet.Close(); closeErr != nil {
	// 		err = errors.Join(err, closeErr)
	// 	}
	// }()

	// buf := make([]byte, 32)
	// _, _ = rand.Read(buf)

	// sig, err := wal.SignText(acct, []byte(hex.EncodeToString(buf)))
	// if err != nil {
	// 	return nil, err
	// }

	// fmt.Println("Signature:", hex.EncodeToString(sig))

	return &http.Client{
		Transport: NewX402BuyerTransport(http.DefaultTransport, wal, acct),
	}, nil
}
