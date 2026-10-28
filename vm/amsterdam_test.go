package vm

import (
	"fmt"
	"github.com/Giulio2002/gevm/spec"
	"github.com/holiman/uint256"
	"testing"
)

// EIP-8024 vectors from execution-specs tests@v21.0.2. Exercise both
// generated dispatchers, not just the immediate decoding helpers.
func TestAmsterdamStackImmediates(t *testing.T) {
	for _, trace := range []bool{false, true} {
		for _, tc := range []struct {
			op, imm byte
			n, m    int
		}{
			{0xe6, 0, 145, 0}, {0xe6, 90, 235, 0}, {0xe6, 128, 17, 0}, {0xe6, 255, 144, 0},
			{0xe7, 0, 145, 0}, {0xe7, 90, 235, 0}, {0xe7, 128, 17, 0}, {0xe7, 255, 144, 0},
			{0xe8, 0, 9, 16}, {0xe8, 80, 14, 16}, {0xe8, 81, 14, 15}, {0xe8, 128, 1, 16}, {0xe8, 255, 1, 22},
		} {
			t.Run(fmt.Sprintf("trace=%v/%x/%x", trace, tc.op, tc.imm), func(t *testing.T) {
				in := NewInterpreter(NewMemory(), NewBytecode([]byte{tc.op, tc.imm, 0}), Inputs{}, false, spec.Amsterdam, 3)
				for i := 1; i <= 256; i++ {
					in.Stack.Push(*uint256.NewInt(uint64(i)))
				}
				want := append([]uint256.Int(nil), in.Stack.Data()...)
				switch tc.op {
				case 0xe6:
					want = append(want, want[256-tc.n])
				case 0xe7:
					want[255], want[255-tc.n] = want[255-tc.n], want[255]
				case 0xe8:
					want[255-tc.n], want[255-tc.m] = want[255-tc.m], want[255-tc.n]
				}
				var r Runner = DefaultRunner{}
				if trace {
					r = NewTracingRunner(opcodeHooks(), spec.Amsterdam)
				}
				r.Run(in, nil)
				if in.HaltResult != InstructionResultStop || in.Gas.Remaining() != 0 {
					t.Fatalf("halt=%v gas=%d", in.HaltResult, in.Gas.Remaining())
				}
				if in.Stack.Len() != len(want) {
					t.Fatalf("len=%d want %d", in.Stack.Len(), len(want))
				}
				for i, v := range want {
					if in.Stack.data[i] != v {
						t.Fatalf("stack[%d]=%v want %v", i, in.Stack.data[i], v)
					}
				}
			})
		}
	}
}
