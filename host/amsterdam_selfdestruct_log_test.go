package host

import (
	"github.com/Giulio2002/gevm/spec"
	"github.com/Giulio2002/gevm/state"
	"github.com/Giulio2002/gevm/types"
	"testing"
)

func TestAmsterdamSelfdestructTransferLog(t *testing.T) {
	for _, target := range []byte{0x20, 0x21} {
		db := newMockDB()
		db.accounts[addr(1)] = &state.AccountInfo{Balance: u(1_000_000_000), CodeHash: types.KeccakEmpty}
		code := []byte{0x60, target, 0xff}
		db.accounts[addr(0x20)] = &state.AccountInfo{Balance: u(42), Code: code, CodeHash: types.Keccak256(code)}
		e := makeEvm(db, spec.Amsterdam, BlockEnv{})
		tx := simpleTx(addr(1), addr(0x20), 0, 1_000_000, 1)
		r := e.Transact(&tx)
		e.ReleaseEvm()
		want := 0
		if target != 0x20 {
			want = 1
		}
		if !r.IsSuccess() || len(r.Logs) != want {
			t.Fatalf("target %x result %+v", target, r)
		}
		if want == 1 && (r.Logs[0].Topics[1][31] != 0x20 || r.Logs[0].Topics[2][31] != target || r.Logs[0].Data[31] != 42) {
			t.Fatal("wrong selfdestruct transfer")
		}
	}
}
