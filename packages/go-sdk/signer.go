package gosdk

import (
	"context"
	"errors"
)

// Signer signs transaction envelopes. It is an interface so callers can back it
// with a local keypair, an HSM, or a remote signing service; the client never
// holds a raw secret key.
type Signer interface {
	// Address returns the signer's account address (a G... strkey).
	Address() string
	// SignEnvelope signs a base64-encoded transaction envelope XDR for the
	// given network passphrase and returns the signed envelope, also base64.
	SignEnvelope(ctx context.Context, envelopeXDR string, networkPassphrase string) (string, error)
}

// ErrSignerAddress is returned when a Signer reports an address that is not a
// valid Stellar account strkey.
var ErrSignerAddress = errors.New("signer address is not a valid account address")

// SignerFunc adapts a function to the Signer interface, pairing it with a fixed
// address.
type SignerFunc struct {
	// Addr is the account address the signing function signs for.
	Addr string
	// Sign performs the signing.
	Sign func(ctx context.Context, envelopeXDR string, networkPassphrase string) (string, error)
}

// Address returns the configured address.
func (s SignerFunc) Address() string { return s.Addr }

// SignEnvelope calls the configured signing function.
func (s SignerFunc) SignEnvelope(ctx context.Context, envelopeXDR string, networkPassphrase string) (string, error) {
	if s.Sign == nil {
		return "", ErrNoSigner
	}
	return s.Sign(ctx, envelopeXDR, networkPassphrase)
}

// ValidateSigner checks that a Signer reports a valid Stellar account address.
func ValidateSigner(s Signer) error {
	if s == nil {
		return ErrNoSigner
	}
	if !IsAccountAddress(s.Address()) {
		return ErrSignerAddress
	}
	return nil
}

// IsAccountAddress reports whether addr is a valid Stellar account strkey.
//
// Validation is delegated to DecodeStrkey so the helper verifies the version
// byte, base32 encoding, payload length, and CRC16-XModem checksum.
func IsAccountAddress(addr string) bool {
	_, kind, err := DecodeStrkey(addr)
	return err == nil && kind == StrkeyAccount
}

// IsContractAddress reports whether addr is a valid Soroban contract strkey.
//
// Validation is delegated to DecodeStrkey so the helper verifies the version
// byte, base32 encoding, payload length, and CRC16-XModem checksum.
func IsContractAddress(addr string) bool {
	_, kind, err := DecodeStrkey(addr)
	return err == nil && kind == StrkeyContract
}
