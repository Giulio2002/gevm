package host

import (
	"github.com/Giulio2002/gevm/spec"
	"github.com/Giulio2002/gevm/state"
	"github.com/Giulio2002/gevm/types"
	"testing"
)

// Values follow execution-specs e27f097, transactions.calculate_intrinsic_cost.
func TestAmsterdamChildReservoir(t *testing.T) {
	db := newMockDB()
	db.accounts[addr(1)] = &state.AccountInfo{Balance: u(1_000_000_000), CodeHash: types.KeccakEmpty}
	// CALL with 30,000 execution gas: state creation must use the reservoir.
	db.accounts[addr(32)] = &state.AccountInfo{Nonce: 1, Code: []byte{0x5f, 0x5f, 0x5f, 0x5f, 0x5f, 0x60, 33, 0x61, 0x75, 0x30, 0xf1, 0x5f, 0x52, 0x60, 32, 0x5f, 0xf3}}
	db.accounts[addr(33)] = &state.AccountInfo{Nonce: 1, Code: []byte{0x60, 1, 0x5f, 0x55, 0}}
	e := makeEvm(db, spec.Amsterdam, BlockEnv{})
	tx := simpleTx(addr(1), addr(32), 0, 20_000_000, 1)
	got := e.Transact(&tx)
	if len(got.Output) != 32 || got.Output[31] != 1 {
		t.Fatalf("child failed: %+v", got)
	}
}

func TestAmsterdamStorageStateGas(t *testing.T) {
	for _, tc := range []struct {
		name string
		tail []byte
		used uint64
		kind ResultKind
	}{
		{"set", []byte{0}, 15000 + 6 + 2100 + 10000 + 64*1530, ResultSuccess},
		{"revert", []byte{0x5f, 0x5f, 0xfd}, 15000 + 6 + 2100 + 10000 + 4, ResultRevert},
		{"invalid", []byte{0xfe}, spec.TxGasLimitCap, ResultHalt},
		{"clear", []byte{0x5f, 0x5f, 0x55}, (15000 + 6 + 2100 + 10000 + 4 + 100) * 4 / 5, ResultSuccess},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db := newMockDB()
			db.accounts[addr(1)] = &state.AccountInfo{Balance: u(1_000_000_000), CodeHash: types.KeccakEmpty}
			db.accounts[addr(32)] = &state.AccountInfo{Nonce: 1, Code: append([]byte{0x60, 1, 0x60, 0, 0x55}, tc.tail...)}
			e := makeEvm(db, spec.Amsterdam, BlockEnv{})
			tx := simpleTx(addr(1), addr(32), 0, 20_000_000, 1)
			got := e.Transact(&tx)
			if got.Kind != tc.kind || got.GasUsed != tc.used {
				t.Fatalf("got %+v want kind %v gas %d", got, tc.kind, tc.used)
			}
		})
	}
}

func TestAmsterdamExecutionGrant(t *testing.T) {
	for _, tc := range []struct {
		name string
		code []byte
		used uint64
		kind ResultKind
	}{
		{"stop", []byte{0}, 15000, ResultSuccess},
		{"invalid", []byte{0xfe}, spec.TxGasLimitCap, ResultHalt},
		{"revert", []byte{0x5f, 0x5f, 0xfd}, 15004, ResultRevert},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db := newMockDB()
			db.accounts[addr(1)] = &state.AccountInfo{Balance: u(1_000_000_000), CodeHash: types.KeccakEmpty}
			db.accounts[addr(32)] = &state.AccountInfo{Nonce: 1, Code: tc.code}
			e := makeEvm(db, spec.Amsterdam, BlockEnv{})
			tx := simpleTx(addr(1), addr(32), 0, 20_000_000, 1)
			got := e.Transact(&tx)
			if got.Kind != tc.kind || got.GasUsed != tc.used {
				t.Fatalf("got %+v want kind %v gas %d", got, tc.kind, tc.used)
			}
		})
	}
}

func TestAmsterdamIntrinsicDecomposition(t *testing.T) {
	e := makeEvm(newMockDB(), spec.Amsterdam, BlockEnv{})
	tx := simpleTx(addr(1), addr(32), 0, 100000, 1)
	for _, tc := range []struct {
		name           string
		self           bool
		value          uint64
		create         bool
		initial, floor uint64
	}{
		{"call", false, 0, false, 15000, 15000},
		{"value", false, 1, false, 21000, 21000},
		{"self", true, 1, false, 12000, 12000},
		{"create", false, 0, true, 24000, 24000},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tx.To = addr(32)
			if tc.self {
				tx.To = tx.Caller
			}
			tx.Value = u(tc.value)
			tx.Kind = TxKindCall
			if tc.create {
				tx.Kind = TxKindCreate
			}
			got := e.calcIntrinsicGas(&tx)
			if got.InitialGas != tc.initial || got.FloorGas != tc.floor {
				t.Fatalf("got %+v want initial=%d floor=%d", got, tc.initial, tc.floor)
			}
		})
	}
	tx.Kind = TxKindCall
	tx.To = addr(32)
	tx.Value = u(0)
	tx.Input = []byte{0, 1}
	got := e.calcIntrinsicGas(&tx)
	if got.InitialGas != 15020 || got.FloorGas != 15128 {
		t.Fatalf("calldata got %+v", got)
	}
}
