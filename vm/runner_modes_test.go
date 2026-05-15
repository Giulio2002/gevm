package vm

import (
	"testing"

	"github.com/Giulio2002/gevm/opcode"
	"github.com/Giulio2002/gevm/spec"
)

func runWithRunner(runner Runner, code []byte, gasLimit uint64) *Interpreter {
	interp := NewInterpreter(NewMemory(), NewBytecode(code), Inputs{}, false, spec.Prague, gasLimit)
	runner.Run(interp, nil)
	return interp
}

func TestDefaultRunnerMatchesPlainRunnerForSimpleBlock(t *testing.T) {
	code := []byte{byte(opcode.PUSH1), 1, byte(opcode.PUSH1), 2, byte(opcode.ADD), byte(opcode.STOP)}

	fast := runWithRunner(DefaultRunner{}, code, 100)
	plain := runWithRunner(PlainRunner{}, code, 100)

	if fast.HaltResult != plain.HaltResult {
		t.Fatalf("halt mismatch: fast=%v plain=%v", fast.HaltResult, plain.HaltResult)
	}
	if fast.Gas.Remaining() != plain.Gas.Remaining() {
		t.Fatalf("gas mismatch: fast=%d plain=%d", fast.Gas.Remaining(), plain.Gas.Remaining())
	}
	if fast.StackLen() != 1 || plain.StackLen() != 1 {
		t.Fatalf("stack len mismatch: fast=%d plain=%d", fast.StackLen(), plain.StackLen())
	}
	if fast.Stack.data[0].Uint64() != 3 || plain.Stack.data[0].Uint64() != 3 {
		t.Fatalf("stack value mismatch: fast=%d plain=%d", fast.Stack.data[0].Uint64(), plain.Stack.data[0].Uint64())
	}
}

func TestDefaultRunnerMatchesPlainRunnerFailures(t *testing.T) {
	for _, tc := range []struct {
		name string
		code []byte
		gas  uint64
	}{
		{name: "underflow", code: []byte{byte(opcode.ADD)}, gas: 100},
		{name: "oog-before-underflow", code: []byte{byte(opcode.ADD)}, gas: 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fast := runWithRunner(DefaultRunner{}, tc.code, tc.gas)
			plain := runWithRunner(PlainRunner{}, tc.code, tc.gas)
			if fast.HaltResult != plain.HaltResult {
				t.Fatalf("halt mismatch: fast=%v plain=%v", fast.HaltResult, plain.HaltResult)
			}
			if fast.Gas.Remaining() != plain.Gas.Remaining() {
				t.Fatalf("gas mismatch: fast=%d plain=%d", fast.Gas.Remaining(), plain.Gas.Remaining())
			}
		})
	}
}

func TestDefaultRunnerGasOpcodeMatchesPlainRunner(t *testing.T) {
	code := []byte{byte(opcode.GAS), byte(opcode.STOP)}

	fast := runWithRunner(DefaultRunner{}, code, 10)
	plain := runWithRunner(PlainRunner{}, code, 10)

	if fast.HaltResult != plain.HaltResult {
		t.Fatalf("halt mismatch: fast=%v plain=%v", fast.HaltResult, plain.HaltResult)
	}
	if fast.StackLen() != 1 || plain.StackLen() != 1 {
		t.Fatalf("stack len mismatch: fast=%d plain=%d", fast.StackLen(), plain.StackLen())
	}
	if fast.Stack.data[0] != plain.Stack.data[0] {
		t.Fatalf("GAS opcode mismatch: fast=%d plain=%d", fast.Stack.data[0].Uint64(), plain.Stack.data[0].Uint64())
	}
}

func TestTracingRunnerWithoutOpcodeHookUsesDefaultPath(t *testing.T) {
	code := []byte{byte(opcode.PUSH1), 1, byte(opcode.PUSH1), 2, byte(opcode.ADD), byte(opcode.STOP)}
	hooks := &Hooks{OnExit: func(int, []byte, uint64, error, bool) {}}

	fast := runWithRunner(DefaultRunner{}, code, 100)
	tracing := runWithRunner(NewTracingRunner(hooks, spec.Prague), code, 100)

	if fast.HaltResult != tracing.HaltResult {
		t.Fatalf("halt mismatch: fast=%v tracing=%v", fast.HaltResult, tracing.HaltResult)
	}
	if fast.Gas.Remaining() != tracing.Gas.Remaining() {
		t.Fatalf("gas mismatch: fast=%d tracing=%d", fast.Gas.Remaining(), tracing.Gas.Remaining())
	}
	if fast.StackLen() != tracing.StackLen() {
		t.Fatalf("stack len mismatch: fast=%d tracing=%d", fast.StackLen(), tracing.StackLen())
	}
}
