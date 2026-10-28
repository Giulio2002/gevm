package state

import (
	"fmt"
	"testing"

	"github.com/Giulio2002/gevm/spec"
	"github.com/Giulio2002/gevm/types"
)

func TestSelfdestructRollbackBalance(t *testing.T) {
	for _, fork := range []spec.ForkID{spec.Shanghai, spec.Osaka, spec.Amsterdam} {
		for _, created := range []bool{false, true} {
			for _, self := range []bool{false, true} {
				for _, repeats := range []int{1, 2} {
					t.Run(fmt.Sprintf("%s/created=%t/self=%t/repeats=%d", fork, created, self, repeats), func(t *testing.T) {
						j := NewJournal(newMockDB())
						j.SetForkID(fork)
						a, b := addr(0x20), addr(0x21)
						acc := NewAccountFromInfo(AccountInfo{Balance: u256(42), Nonce: 1, Code: []byte{0x30, 0xff}, CodeHash: types.Keccak256([]byte{0x30, 0xff})})
						j.State[a] = acc
						j.State[b] = NewAccountFromInfo(AccountInfo{Balance: u256(7)})
						// Creation predates the ancestor checkpoint being reverted.
						if created {
							acc.MarkCreatedLocally()
						}
						target := b
						if self {
							target = a
						}
						ancestor := j.Checkpoint()
						for i := 0; i < repeats; i++ {
							j.Checkpoint()
							result, err := j.Selfdestruct(a, target)
							if err != nil {
								t.Fatal(err)
							}
							if i == 0 && !result.Data.HadValue {
								t.Fatal("SELFDESTRUCT must report its original nonzero balance")
							}
							j.CheckpointCommit()
						}
						j.CheckpointRevert(ancestor)
						if acc.Info.Balance != u256(42) || j.State[b].Info.Balance != u256(7) {
							t.Fatalf("rollback balances: source=%s beneficiary=%s; want 42 and 7", acc.Info.Balance.Dec(), j.State[b].Info.Balance.Dec())
						}
						if acc.IsSelfdestructed() || acc.IsSelfdestructedLocally() || len(j.SelfdestructedAddresses) != 0 || acc.IsCreatedLocally() != created {
							t.Fatal("rollback did not restore destruction/creation status")
						}
						j.CommitTx()
						if acc.Info.Balance != u256(42) || acc.Info.Nonce != 1 || len(acc.Info.Code) != 2 {
							t.Fatal("reverted destruction must not be finalized")
						}
					})
				}
			}
		}
	}
}
