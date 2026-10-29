package host

import (
	"fmt"
	"github.com/Giulio2002/gevm/spec"
	"github.com/Giulio2002/gevm/state"
	"github.com/Giulio2002/gevm/types"
	"github.com/Giulio2002/gevm/vm"
	"testing"
)

func TestAmsterdamTransactionInitcodeLimit(t *testing.T) {
	for _, fork := range []spec.ForkID{spec.Osaka, spec.Amsterdam} {
		for _, size := range []int{49152, 49153, 131072, 131073} {
			t.Run(fmt.Sprintf("%v/%d", fork, size), func(t *testing.T) {
				db := newMockDB()
				db.accounts[addr(1)] = &state.AccountInfo{Balance: u(1_000_000_000), CodeHash: types.KeccakEmpty}
				e := makeEvm(db, fork, BlockEnv{})
				defer e.ReleaseEvm()
				tx := Transaction{Kind: TxKindCreate, Caller: addr(1), Input: make([]byte, size), GasLimit: 10_000_000, GasPrice: u(1)}
				r := e.Transact(&tx)
				limit := 49152
				if fork == spec.Amsterdam {
					limit = 131072
				}
				if size > limit {
					if !r.ValidationError || r.Reason != vm.InstructionResultCreateInitCodeSizeLimit {
						t.Fatalf("want initcode validation error, got %+v", r)
					}
					return
				}
				if !r.IsSuccess() {
					t.Fatalf("want success, got %+v", r)
				}
			})
		}
	}
}
