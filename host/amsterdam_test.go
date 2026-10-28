package host

import (
	"fmt"
	"github.com/Giulio2002/gevm/spec"
	"github.com/Giulio2002/gevm/state"
	"github.com/Giulio2002/gevm/types"
	"github.com/Giulio2002/gevm/vm"
	"testing"
)

func TestAmsterdamRuntimeCodeLimit(t *testing.T) {
	for _, fork := range []spec.ForkID{spec.Osaka, spec.Amsterdam} {
		for _, size := range []int{24576, 24577, 65536, 65537} {
			t.Run(fmt.Sprintf("%v/%d", fork, size), func(t *testing.T) {
				db := newMockDB()
				db.accounts[addr(1)] = &state.AccountInfo{Balance: u(1000), CodeHash: types.KeccakEmpty}
				h, _ := makeHandler(db, fork)
				code := []byte{0x62, byte(size >> 16), byte(size >> 8), byte(size), 0x5f, 0xf3}
				res := execFrame(h, vm.NewFrameInputCreate(vm.CreateInputs{Caller: addr(1), Scheme: vm.NewCreateSchemeCreate(), InitCode: code, GasLimit: 200_000_000}), 0, h.RootMemory)
				limit := 24576
				if fork == spec.Amsterdam {
					limit = 65536
				}
				if size > limit {
					if res.Create.Result.Result != vm.InstructionResultCreateContractSizeLimit {
						t.Fatalf("want size limit, got %v", res.Create.Result.Result)
					}
					return
				}
				if !res.Create.Result.Result.IsOk() || res.Create.Address == nil {
					t.Fatalf("creation failed: %v", res.Create.Result.Result)
				}
			})
		}
	}
}
