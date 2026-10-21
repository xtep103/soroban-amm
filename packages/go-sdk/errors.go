package gosdk

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
)

// Contract error sentinels for contracts/amm's AmmError enum. The discriminants
// are documented in docs/error-codes.md; callers match with errors.Is rather
// than string comparison so the RPC wire format can change without breaking
// them.
var (
	// ErrAlreadyInitialized is AmmError::AlreadyInitialized (1).
	ErrAlreadyInitialized = errors.New("pool already initialized")
	// ErrInvalidFeeBps is AmmError::InvalidFeeBps (2).
	ErrInvalidFeeBps = errors.New("invalid fee bps")
	// ErrInsufficientShares is AmmError::InsufficientShares (3).
	ErrInsufficientShares = errors.New("insufficient shares")
	// ErrDeadlineExceeded is AmmError::DeadlineExceeded (4).
	ErrDeadlineExceeded = errors.New("deadline exceeded")
	// ErrSlippageExceeded is AmmError::SlippageExceeded (5).
	ErrSlippageExceeded = errors.New("slippage exceeded")
	// ErrPaused is AmmError::Paused (6).
	ErrPaused = errors.New("pool is paused")
	// ErrUnauthorized is AmmError::Unauthorized (7).
	ErrUnauthorized = errors.New("unauthorized")
	// ErrZeroAmount is AmmError::ZeroAmount (8).
	ErrZeroAmount = errors.New("zero amount")
	// ErrInvalidToken is AmmError::InvalidToken (9).
	ErrInvalidToken = errors.New("invalid token")
	// ErrEmptyPool is AmmError::EmptyPool (10).
	ErrEmptyPool = errors.New("empty pool")
	// ErrInsufficientLiquidity is AmmError::InsufficientLiquidity (11).
	ErrInsufficientLiquidity = errors.New("insufficient liquidity")
	// ErrNoPendingAdmin is AmmError::NoPendingAdmin (12).
	ErrNoPendingAdmin = errors.New("no pending admin")
	// ErrWrongAdmin is AmmError::WrongAdmin (13).
	ErrWrongAdmin = errors.New("wrong admin")
	// ErrReentrant is AmmError::Reentrant (14).
	ErrReentrant = errors.New("reentrant call")
	// ErrCircuitBreaker is AmmError::CircuitBreaker (15).
	ErrCircuitBreaker = errors.New("circuit breaker tripped")
)

// Client-side sentinels, distinct from contract errors.
var (
	// ErrInvalidConfig is returned by NewClient for a config that cannot
	// produce a usable client.
	ErrInvalidConfig = errors.New("invalid client config")
	// ErrNoSigner is returned when a write method is called on a client
	// constructed without a Signer.
	ErrNoSigner = errors.New("no signer configured")
	// ErrDeadlineRequired is returned when a write method is given a zero
	// deadline. Deadlines are a safety parameter and are never defaulted.
	ErrDeadlineRequired = errors.New("deadline required")
	// ErrSlippageRequired is returned when a write method is given a nil or
	// negative slippage bound.
	ErrSlippageRequired = errors.New("slippage bound required")
	// ErrTransactionFailed is returned when a submitted transaction reaches a
	// terminal non-success status.
	ErrTransactionFailed = errors.New("transaction failed")
	// ErrRPC is returned when the RPC endpoint reports a JSON-RPC error.
	ErrRPC = errors.New("rpc error")
)

// ammErrorsByCode maps AmmError discriminants to sentinels.
var ammErrorsByCode = map[int]error{
	1:  ErrAlreadyInitialized,
	2:  ErrInvalidFeeBps,
	3:  ErrInsufficientShares,
	4:  ErrDeadlineExceeded,
	5:  ErrSlippageExceeded,
	6:  ErrPaused,
	7:  ErrUnauthorized,
	8:  ErrZeroAmount,
	9:  ErrInvalidToken,
	10: ErrEmptyPool,
	11: ErrInsufficientLiquidity,
	12: ErrNoPendingAdmin,
	13: ErrWrongAdmin,
	14: ErrReentrant,
	15: ErrCircuitBreaker,
}

// ContractError is a decoded contract error. It wraps the matching sentinel so
// errors.Is works, and keeps the discriminant and the raw RPC text for logging.
type ContractError struct {
	// Code is the on-chain discriminant, or 0 if it could not be extracted.
	Code int
	// Raw is the unmodified error text returned by the RPC endpoint.
	Raw string

	sentinel error
}

// Error implements the error interface.
func (e *ContractError) Error() string {
	if e.sentinel != nil {
		return fmt.Sprintf("contract error %d: %s", e.Code, e.sentinel.Error())
	}
	return fmt.Sprintf("contract error: %s", e.Raw)
}

// Unwrap exposes the sentinel so errors.Is(err, ErrSlippageExceeded) matches.
func (e *ContractError) Unwrap() error { return e.sentinel }

// contractErrPattern matches the Error(Contract, #N) form Soroban RPC uses to
// report a contract-returned error.
var contractErrPattern = regexp.MustCompile(`Error\s*\(\s*Contract\s*,\s*#(\d+)\s*\)`)

// DecodeContractError converts an RPC error string into a typed *ContractError.
// An unrecognised discriminant still yields a *ContractError so the caller
// keeps the code and raw text; only the sentinel is absent.
func DecodeContractError(raw string) *ContractError {
	ce := &ContractError{Raw: raw}

	m := contractErrPattern.FindStringSubmatch(raw)
	if m == nil {
		return ce
	}

	code, err := strconv.Atoi(m[1])
	if err != nil {
		return ce
	}
	ce.Code = code
	ce.sentinel = ammErrorsByCode[code]
	return ce
}
