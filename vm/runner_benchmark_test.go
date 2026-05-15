package vm

import (
	"testing"

	"github.com/Giulio2002/gevm/opcode"
	"github.com/Giulio2002/gevm/spec"
)

var (
	benchmarkRunnerGasSink uint64
	benchmarkOpcodeSink    uint64
)

func BenchmarkRunnerStraightLine(b *testing.B) {
	code := makeStraightLineBenchCode(256)
	benchmarkRunnerModes(b, code)
}

func BenchmarkRunnerBlockBoundaries(b *testing.B) {
	code := makeBlockBoundaryBenchCode(256)
	benchmarkRunnerModes(b, code)
}

func benchmarkRunnerModes(b *testing.B, code []byte) {
	b.Helper()

	b.Run("Default", func(b *testing.B) {
		benchmarkRunner(b, DefaultRunner{}, code)
	})
	b.Run("TracingNoOpcodeHook", func(b *testing.B) {
		hooks := &Hooks{OnExit: func(int, []byte, uint64, error, bool) {}}
		benchmarkRunner(b, NewTracingRunner(hooks, spec.Prague), code)
	})
	b.Run("TracingOpcodeHook", func(b *testing.B) {
		hooks := &Hooks{
			OnOpcode: func(uint64, byte, uint64, uint64, OpContext, []byte, int, error) {
				benchmarkOpcodeSink++
			},
		}
		benchmarkRunner(b, NewTracingRunner(hooks, spec.Prague), code)
	})
}

func benchmarkRunner(b *testing.B, runner Runner, code []byte) {
	b.Helper()

	const gasLimit = uint64(100_000_000)
	bytecode := NewBytecode(code)
	interp := NewInterpreter(NewMemory(), bytecode, Inputs{}, false, spec.Prague, gasLimit)

	resetBenchmarkInterpreter(interp, gasLimit)
	runner.Run(interp, nil)
	if interp.HaltResult != InstructionResultStop {
		b.Fatalf("warmup halt result: got %v, want %v", interp.HaltResult, InstructionResultStop)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resetBenchmarkInterpreter(interp, gasLimit)
		runner.Run(interp, nil)
		if interp.HaltResult != InstructionResultStop {
			b.Fatalf("halt result: got %v, want %v", interp.HaltResult, InstructionResultStop)
		}
	}
	b.StopTimer()

	benchmarkRunnerGasSink += interp.Gas.Remaining() + uint64(interp.StackLen())
}

func resetBenchmarkInterpreter(interp *Interpreter, gasLimit uint64) {
	interp.Bytecode.pc = 0
	interp.Bytecode.running = true
	interp.Gas = NewGas(gasLimit)
	interp.Stack.Clear()
	interp.ReturnData = nil
	interp.Memory.Reset()
	interp.HasAction = false
	interp.HaltResult = 0
}

func makeStraightLineBenchCode(repetitions int) []byte {
	code := make([]byte, 0, repetitions*6+1)
	for i := 0; i < repetitions; i++ {
		code = append(code,
			byte(opcode.PUSH1), byte(i),
			byte(opcode.PUSH1), byte(i+1),
			byte(opcode.ADD),
			byte(opcode.POP),
		)
	}
	code = append(code, byte(opcode.STOP))
	return code
}

func makeBlockBoundaryBenchCode(repetitions int) []byte {
	code := make([]byte, 0, repetitions*4+1)
	for i := 0; i < repetitions; i++ {
		code = append(code,
			byte(opcode.PUSH1), byte(i),
			byte(opcode.POP),
			byte(opcode.GAS),
			byte(opcode.POP),
		)
	}
	code = append(code, byte(opcode.STOP))
	return code
}
