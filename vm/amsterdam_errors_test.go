package vm

import (
	"github.com/Giulio2002/gevm/spec"
	"testing"
)

func TestAmsterdamStackUnderflow(t *testing.T) {
	for _, op := range []byte{0xe6, 0xe7, 0xe8} {
		for _, trace := range []bool{false, true} {
			in := NewInterpreter(NewMemory(), NewBytecode([]byte{op, 128}), Inputs{}, false, spec.Amsterdam, 3)
			var r Runner = DefaultRunner{}
			if trace {
				r = NewTracingRunner(opcodeHooks(), spec.Amsterdam)
			}
			r.Run(in, nil)
			if in.HaltResult != InstructionResultStackUnderflow {
				t.Errorf("op=%x trace=%v: %v", op, trace, in.HaltResult)
			}
		}
	}
}
