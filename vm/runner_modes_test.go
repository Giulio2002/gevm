package vm

import (
	"reflect"
	"testing"

	"github.com/Giulio2002/gevm/opcode"
	"github.com/Giulio2002/gevm/spec"
)

func runWithRunner(runner Runner, code []byte, gasLimit uint64) *Interpreter {
	interp := NewInterpreter(NewMemory(), NewBytecode(code), Inputs{}, false, spec.Prague, gasLimit)
	runner.Run(interp, nil)
	return interp
}

func TestDefaultRunnerMatchesTracingRunnerForSimpleBlock(t *testing.T) {
	code := []byte{byte(opcode.PUSH1), 1, byte(opcode.PUSH1), 2, byte(opcode.ADD), byte(opcode.STOP)}

	fast := runWithRunner(DefaultRunner{}, code, 100)
	trace := runWithRunner(NewTracingRunner(opcodeHooks(), spec.Prague), code, 100)

	if fast.HaltResult != trace.HaltResult {
		t.Fatalf("halt mismatch: fast=%v trace=%v", fast.HaltResult, trace.HaltResult)
	}
	if fast.Gas.Remaining() != trace.Gas.Remaining() {
		t.Fatalf("gas mismatch: fast=%d trace=%d", fast.Gas.Remaining(), trace.Gas.Remaining())
	}
	if fast.StackLen() != 1 || trace.StackLen() != 1 {
		t.Fatalf("stack len mismatch: fast=%d trace=%d", fast.StackLen(), trace.StackLen())
	}
	if fast.Stack.data[0].Uint64() != 3 || trace.Stack.data[0].Uint64() != 3 {
		t.Fatalf("stack value mismatch: fast=%d trace=%d", fast.Stack.data[0].Uint64(), trace.Stack.data[0].Uint64())
	}
}

func TestDefaultRunnerMatchesTracingRunnerFailures(t *testing.T) {
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
			trace := runWithRunner(NewTracingRunner(opcodeHooks(), spec.Prague), tc.code, tc.gas)
			if fast.HaltResult != trace.HaltResult {
				t.Fatalf("halt mismatch: fast=%v trace=%v", fast.HaltResult, trace.HaltResult)
			}
			if fast.Gas.Remaining() != trace.Gas.Remaining() {
				t.Fatalf("gas mismatch: fast=%d trace=%d", fast.Gas.Remaining(), trace.Gas.Remaining())
			}
		})
	}
}

func TestDefaultRunnerGasOpcodeMatchesTracingRunner(t *testing.T) {
	code := []byte{byte(opcode.GAS), byte(opcode.STOP)}

	fast := runWithRunner(DefaultRunner{}, code, 10)
	trace := runWithRunner(NewTracingRunner(opcodeHooks(), spec.Prague), code, 10)

	if fast.HaltResult != trace.HaltResult {
		t.Fatalf("halt mismatch: fast=%v trace=%v", fast.HaltResult, trace.HaltResult)
	}
	if fast.StackLen() != 1 || trace.StackLen() != 1 {
		t.Fatalf("stack len mismatch: fast=%d trace=%d", fast.StackLen(), trace.StackLen())
	}
	if fast.Stack.data[0] != trace.Stack.data[0] {
		t.Fatalf("GAS opcode mismatch: fast=%d trace=%d", fast.Stack.data[0].Uint64(), trace.Stack.data[0].Uint64())
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

func TestTracingRunnerReportsUnbatchedOpcodeGas(t *testing.T) {
	code := []byte{byte(opcode.PUSH1), 1, byte(opcode.PUSH1), 2, byte(opcode.ADD), byte(opcode.STOP)}
	var got []uint64
	hooks := &Hooks{
		OnOpcode: func(_ uint64, _ byte, gas uint64, _ uint64, _ OpContext, _ []byte, _ int, _ error) {
			got = append(got, gas)
		},
	}

	trace := runWithRunner(NewTracingRunner(hooks, spec.Prague), code, 100)
	if trace.HaltResult != InstructionResultStop {
		t.Fatalf("halt result: got %v, want %v", trace.HaltResult, InstructionResultStop)
	}

	want := []uint64{100, 97, 94, 91}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("opcode gas mismatch: got %v, want %v", got, want)
	}
}

func opcodeHooks() *Hooks {
	return &Hooks{
		OnOpcode: func(uint64, byte, uint64, uint64, OpContext, []byte, int, error) {},
	}
}
