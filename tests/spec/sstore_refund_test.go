package spec

import (
	"github.com/holiman/uint256"
	"testing"

	"github.com/Giulio2002/gevm/host"
	gevmspec "github.com/Giulio2002/gevm/spec"
	"github.com/Giulio2002/gevm/state"
	"github.com/Giulio2002/gevm/types"
)

// TestSstoreRefundBasic verifies that SSTORE produces the correct refund
// when clearing a storage slot (non-zero → zero).
func TestSstoreRefundBasic(t *testing.T) {
	// Bytecode: PUSH1 0x00 PUSH1 0x00 SSTORE STOP
	code := types.Bytes{0x60, 0x00, 0x60, 0x00, 0x55, 0x00}

	caller := types.AddressFrom([]byte{0x01})
	target := types.AddressFrom([]byte{0xAA})

	db := NewMemDB()
	db.InsertAccount(caller, state.AccountInfo{
		Balance:  *uint256.NewInt(1_000_000_000),
		Nonce:    0,
		CodeHash: types.KeccakEmpty,
	}, nil)
	db.InsertAccount(target, state.AccountInfo{
		Balance:  uint256.Int{},
		Nonce:    0,
		CodeHash: types.Keccak256(code),
		Code:     code,
	}, map[uint256.Int]uint256.Int{
		uint256.Int{}: *uint256.NewInt(1),
	})

	forkID := gevmspec.Cancun
	blockEnv := host.BlockEnv{
		Number:   *uint256.NewInt(1),
		GasLimit: *uint256.NewInt(1_000_000),
		BaseFee:  *uint256.NewInt(1),
	}
	cfgEnv := host.CfgEnv{ChainId: *uint256.NewInt(1)}

	evm := host.NewEvm(db, forkID, blockEnv, cfgEnv)
	tx := host.Transaction{
		Kind:     host.TxKindCall,
		TxType:   host.TxTypeLegacy,
		Caller:   caller,
		To:       target,
		Value:    uint256.Int{},
		Input:    nil,
		GasLimit: 100_000,
		GasPrice: *uint256.NewInt(1),
		Nonce:    0,
	}

	result := evm.Transact(&tx)

	if result.ValidationError {
		t.Fatalf("unexpected validation error: reason=%d", result.Reason)
	}
	if result.Kind != host.ResultSuccess {
		t.Fatalf("expected success, got kind=%d reason=%d", result.Kind, result.Reason)
	}
	if result.GasRefund != 4800 {
		t.Errorf("GasRefund = %d, want 4800", result.GasRefund)
	}
}

// TestSstoreRefundAfterSubcall verifies that SSTORE refund works correctly
// when a sub-CALL precedes the SSTORE. This is a regression test for a bug
// where STOP leaked stale sub-call return data as the frame's output, causing
// the parent's RETURNDATASIZE to be wrong.
func TestSstoreRefundAfterSubcall(t *testing.T) {
	caller := types.AddressFrom([]byte{0x01})
	dispatcher := types.AddressFrom([]byte{0xCC})
	callee := types.AddressFrom([]byte{0xBB})

	calleeCode := types.Bytes{0x00} // STOP

	// CALL(gas=50000, addr=0xBB, value=0, argsOff=0, argsLen=0, retOff=0, retLen=0), POP, SSTORE(0, 0), STOP
	dispatcherCode := types.Bytes{
		0x60, 0x00, 0x60, 0x00, 0x60, 0x00, 0x60, 0x00, 0x60, 0x00,
		0x73,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0xBB,
		0x61, 0xC3, 0x50,
		0xF1, 0x50,
		0x60, 0x00, 0x60, 0x00, 0x55, 0x00,
	}

	db := NewMemDB()
	db.InsertAccount(caller, state.AccountInfo{
		Balance:  *uint256.NewInt(1_000_000_000),
		Nonce:    0,
		CodeHash: types.KeccakEmpty,
	}, nil)
	db.InsertAccount(dispatcher, state.AccountInfo{
		Balance:  uint256.Int{},
		Nonce:    0,
		CodeHash: types.Keccak256(dispatcherCode),
		Code:     dispatcherCode,
	}, map[uint256.Int]uint256.Int{
		uint256.Int{}: *uint256.NewInt(1),
	})
	db.InsertAccount(callee, state.AccountInfo{
		Balance:  uint256.Int{},
		Nonce:    0,
		CodeHash: types.Keccak256(calleeCode),
		Code:     calleeCode,
	}, nil)

	evm := host.NewEvm(db, gevmspec.Cancun, host.BlockEnv{
		Number:   *uint256.NewInt(1),
		GasLimit: *uint256.NewInt(1_000_000),
		BaseFee:  *uint256.NewInt(1),
	}, host.CfgEnv{ChainId: *uint256.NewInt(1)})

	result := evm.Transact(&host.Transaction{
		Kind: host.TxKindCall, TxType: host.TxTypeLegacy,
		Caller: caller, To: dispatcher,
		GasLimit: 200_000, GasPrice: *uint256.NewInt(1),
	})

	if result.Kind != host.ResultSuccess {
		t.Fatalf("expected success, got kind=%d reason=%d", result.Kind, result.Reason)
	}
	if result.GasRefund != 4800 {
		t.Errorf("GasRefund = %d, want 4800", result.GasRefund)
	}
}
