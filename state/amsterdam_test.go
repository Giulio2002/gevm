package state

import (
	"github.com/Giulio2002/gevm/spec"
	"github.com/Giulio2002/gevm/types"
	"github.com/holiman/uint256"
	"testing"
)

func TestAmsterdamSelfdestructDoesNotBurn(t *testing.T) {
	for _, fork := range []spec.ForkID{spec.Osaka, spec.Amsterdam} {
		t.Run(fork.String(), func(t *testing.T) {
			j := NewJournal(newMockDB())
			j.SetForkID(fork)
			a := addr(1)
			acc := NewAccountFromInfo(AccountInfo{Balance: u256(42), Nonce: 1, Code: []byte{0x00}, CodeHash: types.Keccak256([]byte{0x00})})
			j.State[a] = acc
			acc.MarkCreatedLocally()
			acc.MarkTouch()
			acc.EnsureStorage()
			acc.Storage[u256(1)] = NewEvmStorageSlot(u256(7), 0)
			if _, err := j.Selfdestruct(a, a); err != nil {
				t.Fatal(err)
			}
			want := uint256.Int{}
			if fork == spec.Amsterdam {
				want = u256(42)
			}
			if acc.Info.Balance != want {
				t.Fatalf("selfdestruct balance=%s want %s", acc.Info.Balance.Hex(), want.Hex())
			}
			j.CommitTx()
			if acc.Info.Balance != want || acc.Info.Nonce != 0 || len(acc.Info.Code) != 0 || len(acc.Storage) != 0 {
				t.Fatalf("final account: %+v", acc)
			}
			if fork == spec.Amsterdam && acc.IsSelfdestructed() {
				t.Fatal("funded account must remain live after clearing code/storage")
			}
		})
	}
}
