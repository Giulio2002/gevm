package vm

import (
	"fmt"
	"github.com/Giulio2002/gevm/spec"
	"github.com/holiman/uint256"
	"testing"
)

func TestAmsterdamCreateInitcodeLimit(t *testing.T) {
	for _, fork := range []spec.ForkID{spec.Osaka, spec.Amsterdam} {
		for _, op := range []byte{0xf0, 0xf5} {
			for _, size := range []int{49152, 49153, 131072, 131073} {
				t.Run(fmt.Sprintf("%v/%x/%d", fork, op, size), func(t *testing.T) {
					in := NewInterpreter(NewMemory(), NewBytecode([]byte{op}), Inputs{}, false, fork, 10_000_000)
					if op == 0xf5 {
						in.Stack.Push(uint256.Int{})
					}
					in.Stack.Push(*uint256.NewInt(uint64(size)))
					in.Stack.Push(uint256.Int{})
					in.Stack.Push(uint256.Int{})
					DefaultRunner{}.Run(in, nil)
					limit := 49152
					if fork == spec.Amsterdam {
						limit = 131072
					}
					if size > limit {
						if in.HaltResult != InstructionResultCreateInitCodeSizeLimit {
							t.Fatalf("want initcode limit, got %v", in.HaltResult)
						}
						return
					}
					if !in.HasAction || len(in.ActionData.Create.InitCode) != size {
						t.Fatalf("want create action, got halt=%v action=%v", in.HaltResult, in.HasAction)
					}
				})
			}
		}
	}
}
