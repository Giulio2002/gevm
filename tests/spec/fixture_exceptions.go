package spec

import (
	"fmt"
	"github.com/Giulio2002/gevm/host"
	"github.com/Giulio2002/gevm/precompiles"
	gevmspec "github.com/Giulio2002/gevm/spec"
	"github.com/Giulio2002/gevm/types"
	"github.com/Giulio2002/gevm/vm"
	"github.com/holiman/uint256"
	"strings"
)

const txExceptionPrefix = "TransactionException."

var supportedExceptions = map[string]bool{}

func init() {
	for _, s := range strings.Fields(`INTRINSIC_GAS_TOO_LOW INTRINSIC_GAS_BELOW_FLOOR_GAS_COST INSUFFICIENT_ACCOUNT_FUNDS INVALID_SIGNATURE_VRS SENDER_NOT_EOA INVALID_CHAINID GAS_LIMIT_EXCEEDS_MAXIMUM TYPE_3_TX_INVALID_BLOB_VERSIONED_HASH INSUFFICIENT_MAX_FEE_PER_GAS INITCODE_SIZE_EXCEEDED NONCE_IS_MAX TYPE_3_TX_BLOB_COUNT_EXCEEDED TYPE_3_TX_ZERO_BLOBS TYPE_3_TX_MAX_BLOB_GAS_ALLOWANCE_EXCEEDED INSUFFICIENT_MAX_FEE_PER_BLOB_GAS PRIORITY_GREATER_THAN_MAX_FEE_PER_GAS NONCE_MISMATCH_TOO_LOW NONCE_MISMATCH_TOO_HIGH GAS_ALLOWANCE_EXCEEDED TYPE_3_TX_CONTRACT_CREATION GASLIMIT_PRICE_PRODUCT_OVERFLOW TYPE_4_TX_CONTRACT_CREATION TYPE_4_EMPTY_AUTHORIZATION_LIST`) {
		supportedExceptions[txExceptionPrefix+s] = true
	}
}
func knownExceptions(expected string) bool {
	for _, s := range strings.Split(expected, "|") {
		if !supportedExceptions[s] {
			return false
		}
	}
	return true
}
func matchesException(expected, actual string) bool {
	for _, s := range strings.Split(expected, "|") {
		if s == txExceptionPrefix+actual {
			return true
		}
	}
	return false
}

// executeExceptionCase never accepts an unrelated error as a rejection match.
// It also checks that validation left the complete state and logs unchanged.
func executeExceptionCase(unit *TestUnit, tc *TestCase, fork gevmspec.ForkID) (err error) {
	defer func() {
		if p := recover(); p != nil {
			err = fmt.Errorf("panic validating rejection: %v", p)
		}
	}()
	caller, err := RecoverAddress(unit.Transaction.SecretKey.V[:])
	if err != nil {
		if unit.Transaction.Sender == nil {
			return err
		}
		caller = unit.Transaction.Sender.V
	}
	tx, err := BuildTransaction(unit, tc, caller)
	if err != nil {
		return err
	}
	db := BuildMemDB(unit.Pre)
	e := newTestEVM(db, fork, BuildBlockEnv(unit, fork), host.CfgEnv{ChainId: unit.ChainId()})
	actual := ""
	if tc.TxBytes != nil {
		actual, err = validateFixtureEnvelope(tc.TxBytes.V, unit.ChainId(), fork)
		if err != nil {
			return err
		}
	}
	if actual == "" {
		result := e.Transact(&tx)
		if !result.ValidationError {
			return fmt.Errorf("expected %s, transaction executed: %v", *tc.ExpectException, result.Reason)
		}
		actual = executionValidationException(result.Reason, &tx, unit)
		if len(result.Logs) != 0 {
			return fmt.Errorf("validation emitted logs")
		}
	}
	if !matchesException(*tc.ExpectException, actual) {
		return fmt.Errorf("expected %s, got %s%s", *tc.ExpectException, txExceptionPrefix, actual)
	}
	root := stateRoot(db, e.Journal, fork)
	if root != tc.Hash.V {
		return fmt.Errorf("rejection state root: got %s want %s", root.Hex(), tc.Hash.V.Hex())
	}
	if LogsRoot(nil) != tc.Logs.V {
		return fmt.Errorf("rejection logs hash mismatch")
	}
	return nil
}

func executionValidationException(reason vm.InstructionResult, tx *host.Transaction, unit *TestUnit) string {
	switch reason {
	case vm.InstructionResultGasFloorTooLow:
		return "INTRINSIC_GAS_BELOW_FLOOR_GAS_COST"
	case vm.InstructionResultBlockGasLimitTooHigh:
		return "GAS_ALLOWANCE_EXCEEDED"
	case vm.InstructionResultOutOfGas:
		return "INTRINSIC_GAS_TOO_LOW"
	case vm.InstructionResultOutOfFunds:
		return "INSUFFICIENT_ACCOUNT_FUNDS"
	case vm.InstructionResultSenderNotEOA:
		return "SENDER_NOT_EOA"
	case vm.InstructionResultGasLimitTooHigh:
		return "GAS_LIMIT_EXCEEDS_MAXIMUM"
	case vm.InstructionResultInvalidBlobVersion:
		return "TYPE_3_TX_INVALID_BLOB_VERSIONED_HASH"
	case vm.InstructionResultGasPriceBelowBaseFee:
		return "INSUFFICIENT_MAX_FEE_PER_GAS"
	case vm.InstructionResultCreateInitCodeSizeLimit:
		return "INITCODE_SIZE_EXCEEDED"
	case vm.InstructionResultNonceOverflow:
		return "NONCE_IS_MAX"
	case vm.InstructionResultTooManyBlobs:
		return "TYPE_3_TX_BLOB_COUNT_EXCEEDED"
	case vm.InstructionResultEmptyBlobs:
		return "TYPE_3_TX_ZERO_BLOBS"
	case vm.InstructionResultBlobGasPriceTooHigh:
		return "INSUFFICIENT_MAX_FEE_PER_BLOB_GAS"
	case vm.InstructionResultPriorityFeeTooHigh:
		return "PRIORITY_GREATER_THAN_MAX_FEE_PER_GAS"
	case vm.InstructionResultNonceMismatch:
		if pre := unit.Pre[HexAddr{V: tx.Caller}]; pre != nil && tx.Nonce < pre.Nonce.V {
			return "NONCE_MISMATCH_TOO_LOW"
		}
		return "NONCE_MISMATCH_TOO_HIGH"
	case vm.InstructionResultCreateNotAllowed:
		if tx.TxType == host.TxTypeEIP7702 {
			return "TYPE_4_TX_CONTRACT_CREATION"
		}
		return "TYPE_3_TX_CONTRACT_CREATION"
	case vm.InstructionResultEmptyAuthorizationList:
		return "TYPE_4_EMPTY_AUTHORIZATION_LIST"
	}
	return "UNCLASSIFIED_" + reason.String()
}

func encodeRlpItem(item RlpItem) []byte {
	if item.Kind == RlpString {
		return RlpEncodeBytes(item.Data)
	}
	children := make([][]byte, len(item.Items))
	for i, c := range item.Items {
		children[i] = encodeRlpItem(c)
	}
	return RlpEncodeList(children)
}

// Validate the actual signed fixture envelope, including recovery. The execution
// host accepts already-recovered transactions, so signatures belong in this layer.
func validateFixtureEnvelope(raw []byte, chain uint256.Int, fork gevmspec.ForkID) (string, error) {
	if len(raw) == 0 {
		return "", fmt.Errorf("missing envelope")
	}
	typed := raw[0] < 0x80
	payload := raw
	if typed {
		payload = raw[1:]
	}
	item, err := RlpDecodeComplete(payload)
	if err != nil {
		return "", err
	}
	n := len(item.Items)
	if n < 9 {
		return "", fmt.Errorf("malformed transaction envelope")
	}
	v256, ve := item.Items[n-3].AsU256()
	v := v256.Uint64()
	r, re := item.Items[n-2].AsU256()
	s, se := item.Items[n-1].AsU256()
	if ve != nil || re != nil || se != nil || r.IsZero() || s.IsZero() || !r.Lt(&secp256k1N) || !s.Lt(&secp256k1N) || (fork.IsEnabledIn(gevmspec.Homestead) && s.Gt(&secp256k1HalfN)) {
		return "INVALID_SIGNATURE_VRS", nil
	}
	signing := item.Items[:n-3]
	recid := byte(v)
	if typed {
		if raw[0] < 1 || raw[0] > 4 {
			return "", fmt.Errorf("unsupported envelope type %d", raw[0])
		}
		if v256.Gt(uint256.NewInt(1)) {
			return "INVALID_SIGNATURE_VRS", nil
		}
		got, err := item.Items[0].AsU256()
		if err != nil {
			return "", err
		}
		if got != chain {
			return "INVALID_CHAINID", nil
		}
	} else {
		if v256.Eq(uint256.NewInt(27)) || v256.Eq(uint256.NewInt(28)) {
			recid = byte(v - 27)
		} else if v256.Gt(uint256.NewInt(34)) {
			var derived uint256.Int
			derived.Sub(&v256, uint256.NewInt(35))
			derived.Rsh(&derived, 1)
			id, overflow := derived.Uint64WithOverflow()
			if overflow {
				return "INVALID_SIGNATURE_VRS", nil
			}
			if chain != *uint256.NewInt(id) {
				return "INVALID_CHAINID", nil
			}
			recid = byte((v - 35) % 2)
			signing = append(append([]RlpItem{}, signing...), RlpItem{Kind: RlpString, Data: uint256.NewInt(id).Bytes()}, RlpItem{Kind: RlpString}, RlpItem{Kind: RlpString})
		} else {
			return "INVALID_SIGNATURE_VRS", nil
		}
	}
	encoded := encodeRlpItem(RlpItem{Kind: RlpList, Items: signing})
	if typed {
		encoded = append([]byte{raw[0]}, encoded...)
	}
	digest := types.Keccak256(encoded)
	var sig [64]byte
	rb, sb := r.Bytes32(), s.Bytes32()
	copy(sig[:32], rb[:])
	copy(sig[32:], sb[:])
	if _, ok := precompiles.Ecrecover(sig, recid, [32]byte(digest)); !ok {
		return "INVALID_SIGNATURE_VRS", nil
	}
	return "", nil
}
