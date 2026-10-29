package spec

import (
	gevmspec "github.com/Giulio2002/gevm/spec"
	"github.com/Giulio2002/gevm/state"
	"github.com/Giulio2002/gevm/types"
	"github.com/holiman/uint256"
	"testing"
)

func TestAmsterdamStateRootPreservesDestructionBalance(t *testing.T) {
	db := NewMemDB()
	j := state.NewJournal(db)
	j.SetForkID(gevmspec.Amsterdam)
	a := types.Address{19: 1}
	info := state.AccountInfo{Balance: *uint256.NewInt(42), Nonce: 1, CodeHash: types.Keccak256([]byte{0x00}), Code: []byte{0x00}}
	acc := state.NewAccountFromInfo(info)
	acc.MarkCreatedLocally()
	acc.MarkTouch()
	j.State[a] = acc
	if _, err := j.Selfdestruct(a, a); err != nil {
		t.Fatal(err)
	}
	expected := NewMemDB()
	expected.InsertAccount(a, state.AccountInfo{Balance: *uint256.NewInt(42), CodeHash: types.KeccakEmpty}, nil)
	want := stateRoot(expected, state.NewJournal(expected), gevmspec.Amsterdam)
	if got := stateRoot(db, j, gevmspec.Amsterdam); got != want {
		t.Fatalf("got %s want %s", got.Hex(), want.Hex())
	}
}
