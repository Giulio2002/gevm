package vm

import (
	"fmt"
	"github.com/Giulio2002/gevm/spec"
	"github.com/holiman/uint256"
	"testing"
)

func TestAmsterdamStackImmediateSpace(t *testing.T) {
	for _, op := range []byte{0xe6, 0xe7, 0xe8} {
		for x := 0; x < 256; x++ {
			for _, trace := range []bool{false, true} {
				t.Run(fmt.Sprintf("%x/%02x/trace=%v", op, x, trace), func(t *testing.T) {
					in := NewInterpreter(NewMemory(), NewBytecode([]byte{op, byte(x)}), Inputs{}, false, spec.Amsterdam, 3)
					for i := 0; i < 256; i++ {
						in.Stack.Push(*uint256.NewInt(uint64(i + 1)))
					}
					want := append([]uint256.Int(nil), in.Stack.Data()...)
					valid := x <= 90 || x >= 128
					if op == 0xe8 {
						valid = x <= 81 || x >= 128
					}
					if valid {
						n := (x + 145) % 256
						switch op {
						case 0xe6:
							want = append(want, want[256-n])
						case 0xe7:
							want[255], want[255-n] = want[255-n], want[255]
						case 0xe8:
							q, r := (x^143)/16, (x^143)%16
							n, m := r+1, 29-q
							if q < r {
								n, m = q+1, r+1
							}
							want[255-n], want[255-m] = want[255-m], want[255-n]
						}
					}
					var runner Runner = DefaultRunner{}
					if trace {
						runner = NewTracingRunner(opcodeHooks(), spec.Amsterdam)
					}
					runner.Run(in, nil)
					if !valid {
						if in.HaltResult != InstructionResultInvalidImmediateEncoding {
							t.Fatalf("invalid immediate: %v", in.HaltResult)
						}
						return
					}
					if in.HaltResult != InstructionResultStop || in.Stack.Len() != len(want) || in.Gas.Remaining() != 0 {
						t.Fatalf("halt=%v len=%d gas=%d", in.HaltResult, in.Stack.Len(), in.Gas.Remaining())
					}
					for i := range want {
						if want[i] != in.Stack.data[i] {
							t.Fatalf("stack[%d] mismatch", i)
						}
					}
				})
			}
		}
	}
}

func TestAmsterdamStackBoundaries(t *testing.T) {
	for _, tc := range []struct {
		name  string
		code  []byte
		fork  spec.ForkID
		depth int
		gas   uint64
		want  InstructionResult
	}{
		{"before-fork", []byte{0xe6, 128}, spec.Osaka, 17, 3, InstructionResultNotActivated},
		{"dupn-overflow", []byte{0xe6, 128}, spec.Amsterdam, 1024, 3, InstructionResultStackOverflow},
		{"swapn-full-stack", []byte{0xe7, 128}, spec.Amsterdam, 1024, 3, InstructionResultStop},
		{"dupn-underflow-boundary", []byte{0xe6, 128}, spec.Amsterdam, 16, 3, InstructionResultStackUnderflow},
		{"swapn-underflow-boundary", []byte{0xe7, 128}, spec.Amsterdam, 17, 3, InstructionResultStackUnderflow},
		{"zero-padded-immediate", []byte{0xe6}, spec.Amsterdam, 145, 3, InstructionResultStop},
		{"gas-one-short", []byte{0xe6, 128}, spec.Amsterdam, 17, 2, InstructionResultOutOfGas},
		{"jump-into-forbidden-immediate", []byte{0x60, 4, 0x56, 0xe6, 0x5b, 0}, spec.Amsterdam, 0, 12, InstructionResultStop},
	} {
		t.Run(tc.name, func(t *testing.T) {
			in := NewInterpreter(NewMemory(), NewBytecode(tc.code), Inputs{}, false, tc.fork, tc.gas)
			for i := 0; i < tc.depth; i++ {
				in.Stack.Push(uint256.Int{})
			}
			DefaultRunner{}.Run(in, nil)
			if in.HaltResult != tc.want {
				t.Fatalf("got %v want %v", in.HaltResult, tc.want)
			}
		})
	}
}
