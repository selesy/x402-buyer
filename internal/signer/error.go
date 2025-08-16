package signer

import "errors"

// ErrAccountNotFound is returned if the account passed in constructor is
// not present in the keystore.
var ErrAccountNotFound = errors.New("account not found in keystore")

// ErrEnvVarNotFound is returned when the environment variable that's
// supposed to contain the private key's hexadecimal value is not present.
var ErrEnvVarNotFound = errors.New("environment variable not found")

// ErrInvalidCurve is returned when the curve is not the Ethereum secp256k1
// curve.
var ErrInvalidCurve = errors.New("curve must be secp256k1 curve from go-ethereum")

// ErrInvalidPoint is returned if the X, Y coordinates of the provided point
// are not on the secp256k1 curve.
var ErrInvalidPoint = errors.New("point coordinates must be on the secp256k1 curve")
