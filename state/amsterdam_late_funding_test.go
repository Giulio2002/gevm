package state

import (
	"github.com/Giulio2002/gevm/spec"
	"github.com/Giulio2002/gevm/types"
	"testing"
)

func TestAmsterdamSelfdestructLateFunding(t *testing.T) {
	for _, revert := range []bool{false, true} {
		db := newMockDB()
		a, b, c := addr(1), addr(2), addr(3)
		db.accounts[a] = &AccountInfo{Balance: u256(42), Nonce: 1, CodeHash: types.Keccak256([]byte{0x00}), Code: []byte{0x00}}
		db.accounts[c] = &AccountInfo{Balance: u256(100), CodeHash: types.KeccakEmpty}
		j := NewJournal(db)
		j.SetForkID(spec.Amsterdam)
		load, err := j.LoadAccount(a)
		if err != nil {
			t.Fatal(err)
		}
		acc := load.Data
		acc.MarkCreatedLocally()
		acc.MarkTouch()
		cp := j.Checkpoint()
		if _, err = j.Selfdestruct(a, b); err != nil {
			t.Fatal(err)
		}
		if e, err := j.Transfer(c, a, u256(19)); e != nil || err != nil {
			t.Fatalf("transfer: %v %v", e, err)
		}
		if revert {
			j.CheckpointRevert(cp)
		} else {
			j.CheckpointCommit()
		}
		j.CommitTx()
		want := u256(19)
		nonce := uint64(0)
		if revert {
			want = u256(42)
			nonce = 1
		}
		if acc.Info.Balance != want || acc.Info.Nonce != nonce || acc.IsSelfdestructed() {
			t.Fatalf("revert=%v: %+v", revert, acc)
		}
		if !revert && (len(acc.Info.Code) != 0 || !acc.IsStorageCleared()) {
			t.Fatal("destroyed account not cleared")
		}
	}
}
