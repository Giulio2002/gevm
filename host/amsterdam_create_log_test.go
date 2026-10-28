package host

import (
	"github.com/Giulio2002/gevm/spec"
	"github.com/Giulio2002/gevm/state"
	"github.com/Giulio2002/gevm/types"
	"testing"
)

func TestAmsterdamCreateTransferLog(t *testing.T) {
	for _, code := range [][]byte{{0}, {0x5f, 0x5f, 0xfd}, {0xfe}} {
		db := newMockDB()
		db.accounts[addr(1)] = &state.AccountInfo{Balance: u(1_000_000_000), CodeHash: types.KeccakEmpty}
		e := makeEvm(db, spec.Amsterdam, BlockEnv{})
		tx := Transaction{Kind: TxKindCreate, Caller: addr(1), Value: u(7), Input: code, GasLimit: 1_000_000, GasPrice: u(1)}
		r := e.Transact(&tx)
		e.ReleaseEvm()
		want := 0
		if code[0] == 0 {
			want = 1
		}
		if len(r.Logs) != want {
			t.Fatalf("code %x logs=%d want %d", code, len(r.Logs), want)
		}
		if want == 1 && (r.CreatedAddr == nil || string(r.Logs[0].Topics[2][12:]) != string(r.CreatedAddr[:])) {
			t.Fatal("wrong creation recipient")
		}
	}
}
