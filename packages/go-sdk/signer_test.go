package gosdk

import (
	"context"
	"errors"
	"net/http"
	"testing"
)

func corruptStrkey(addr string) string {
	corrupt := []byte(addr)
	if corrupt[10] == 'A' {
		corrupt[10] = 'B'
	} else {
		corrupt[10] = 'A'
	}
	return string(corrupt)
}

func TestIsAccountAddressRejectsCorruptedChecksum(t *testing.T) {
	corrupt := corruptStrkey(zeroAccount)

	if IsAccountAddress(corrupt) {
		t.Fatalf("checksum-corrupted account address %q must be rejected", corrupt)
	}
}

func TestIsContractAddressRejectsCorruptedChecksum(t *testing.T) {
	corrupt := corruptStrkey(zeroContract)

	if IsContractAddress(corrupt) {
		t.Fatalf("checksum-corrupted contract address %q must be rejected", corrupt)
	}
}

func TestAddressHelpersAcceptValidAddresses(t *testing.T) {
	if !IsAccountAddress(zeroAccount) {
		t.Fatalf("zeroAccount should be accepted as a valid account address")
	}

	if IsContractAddress(zeroAccount) {
		t.Fatalf("zeroAccount must not be accepted as a contract address")
	}

	if !IsContractAddress(zeroContract) {
		t.Fatalf("zeroContract should be accepted as a valid contract address")
	}

	if IsAccountAddress(zeroContract) {
		t.Fatalf("zeroContract must not be accepted as an account address")
	}
}

func TestNewClientRejectsChecksumCorruptedSourceAccount(t *testing.T) {
	corrupt := corruptStrkey(zeroAccount)

	client, err := NewClient(Config{
		RPCURL:            "http://localhost:8000",
		NetworkPassphrase: NetworkTestnet,
		SourceAccount:     corrupt,
	})
	if !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("expected ErrInvalidConfig, got %v", err)
	}
	if client != nil {
		t.Fatal("NewClient must not return a client for a checksum-corrupted source account")
	}
}

func TestValidateSignerRejectsChecksumCorruptedAddress(t *testing.T) {
	corrupt := corruptStrkey(zeroAccount)

	signer := SignerFunc{
		Addr: corrupt,
		Sign: func(ctx context.Context, envelopeXDR string, networkPassphrase string) (string, error) {
			return envelopeXDR, nil
		},
	}

	err := ValidateSigner(signer)
	if !errors.Is(err, ErrSignerAddress) {
		t.Fatalf("expected ErrSignerAddress, got %v", err)
	}
}

func TestSharesOfRejectsChecksumCorruptedProviderBeforeRPC(t *testing.T) {
	corrupt := corruptStrkey(zeroAccount)

	client, err := NewClient(Config{
		RPCURL:            "http://localhost:8000",
		NetworkPassphrase: NetworkTestnet,
		SourceAccount:     zeroAccount,
	})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	client.WithHTTPClient(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			t.Fatalf("SharesOf must reject the provider before making an HTTP request")
			return nil, nil
		}),
	})

	_, err = client.SharesOf(context.Background(), zeroContract, corrupt)
	if !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("expected ErrInvalidConfig, got %v", err)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
