package api

import "github.com/ethereum/go-ethereum/common"

// A Signer is implemented by types that can produce a ECDSA signature
// of the provided digestHash.
type Signer interface {
	Sign(digestHash []byte) ([]byte, error)
}

// An EVMSigner is a Signer that operates on behalf of an Ethereum account
// and therefore has an address.
type EVMSigner interface {
	Signer

	Address() common.Address
}
