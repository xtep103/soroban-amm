package gosdk

import (
	"encoding/base64"
	"fmt"
)

// InvokeSpec describes the contract invocation an envelope should carry.
type InvokeSpec struct {
	// SourceAccount pays the fee and, for writes, is the signing account.
	SourceAccount string
	// ContractID is the invoked contract's address.
	ContractID string
	// Method is the contract function name.
	Method string
	// Args are the invocation arguments, in declaration order.
	Args []ScValue
	// NetworkPassphrase identifies the network the envelope targets.
	NetworkPassphrase string
	// Sequence is the source account's next sequence number. Simulation
	// ignores it, so it may be left zero for read-only calls.
	Sequence int64
	// Fee is the transaction fee in stroops. Defaults to BaseFee when zero;
	// simulation returns the resource fee that must be added before submission.
	Fee uint32
	// TimeoutSeconds becomes the envelope's absolute time bound offset. Zero
	// means no time bound.
	TimeoutSeconds uint64
	// MinLedgerTime is the lower time bound; normally zero.
	MinLedgerTime uint64
}

// BaseFee is the default per-operation fee in stroops.
const BaseFee uint32 = 100

// XDR constants for the envelope structure this client emits.
const (
	envelopeTypeTx      int32 = 2
	txPreconditionsTime int32 = 1
	txPreconditionsNone int32 = 0
	opTypeInvokeHostFn  int32 = 24
	hostFnTypeInvoke    int32 = 0
	keyTypeEd25519      int32 = 0
	muxedAccountEd25519 int32 = 0
)

// BuildInvokeEnvelope builds a base64-encoded TransactionEnvelope XDR carrying
// a single InvokeHostFunction operation. The envelope carries no signatures;
// a Signer adds those before submission.
func BuildInvokeEnvelope(spec InvokeSpec) (string, error) {
	if !IsAccountAddress(spec.SourceAccount) {
		return "", fmt.Errorf("%w: source account %q", ErrInvalidConfig, spec.SourceAccount)
	}
	if !IsContractAddress(spec.ContractID) {
		return "", fmt.Errorf("%w: contract id %q", ErrInvalidConfig, spec.ContractID)
	}
	if spec.Method == "" {
		return "", fmt.Errorf("%w: empty method name", ErrInvalidConfig)
	}
	if len(spec.Method) > 32 {
		return "", fmt.Errorf("%w: method %q exceeds the 32-character symbol limit", ErrInvalidConfig, spec.Method)
	}

	sourceRaw, kind, err := DecodeStrkey(spec.SourceAccount)
	if err != nil {
		return "", err
	}
	if kind != StrkeyAccount {
		return "", fmt.Errorf("%w: source %q is not an account address", ErrInvalidConfig, spec.SourceAccount)
	}

	fee := spec.Fee
	if fee == 0 {
		fee = BaseFee
	}

	var w xdrWriter

	// TransactionEnvelope: union discriminated on EnvelopeType.
	w.writeInt32(envelopeTypeTx)

	// Transaction.sourceAccount: MuxedAccount union.
	w.writeInt32(muxedAccountEd25519)
	w.writeRaw(sourceRaw)

	w.writeUint32(fee)
	w.writeInt64(spec.Sequence)

	// Transaction.cond: Preconditions union.
	if spec.TimeoutSeconds > 0 || spec.MinLedgerTime > 0 {
		w.writeInt32(txPreconditionsTime)
		w.writeUint64(spec.MinLedgerTime)
		w.writeUint64(spec.TimeoutSeconds)
	} else {
		w.writeInt32(txPreconditionsNone)
	}

	// Transaction.memo: MEMO_NONE.
	w.writeInt32(0)

	// Transaction.operations: exactly one.
	w.writeUint32(1)

	// Operation.sourceAccount: absent, so the transaction source is used.
	w.writeUint32(0)

	// Operation.body: InvokeHostFunction.
	w.writeInt32(opTypeInvokeHostFn)
	w.writeInt32(hostFnTypeInvoke)

	// InvokeContractArgs.contractAddress: SCAddress union.
	contractRaw, kind, err := DecodeStrkey(spec.ContractID)
	if err != nil {
		return "", err
	}
	if kind != StrkeyContract {
		return "", fmt.Errorf("%w: %q is not a contract address", ErrInvalidConfig, spec.ContractID)
	}
	w.writeInt32(scAddressCntrct)
	w.writeRaw(contractRaw)

	// InvokeContractArgs.functionName: SCSymbol.
	w.writeBytes([]byte(spec.Method))

	// InvokeContractArgs.args.
	w.writeUint32(uint32(len(spec.Args)))
	for i, arg := range spec.Args {
		if err := writeScVal(&w, arg); err != nil {
			return "", fmt.Errorf("encoding argument %d of %s: %w", i, spec.Method, err)
		}
	}

	// InvokeHostFunctionOp.auth: empty; simulation fills it in.
	w.writeUint32(0)

	// Transaction.ext: v0. Simulation returns the SorobanTransactionData that
	// replaces this before submission.
	w.writeInt32(0)

	// TransactionV1Envelope.signatures: none yet.
	w.writeUint32(0)

	return base64.StdEncoding.EncodeToString(w.bytes()), nil
}
