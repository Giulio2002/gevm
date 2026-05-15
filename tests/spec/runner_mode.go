package spec

import (
	"os"

	"github.com/Giulio2002/gevm/host"
	gevmspec "github.com/Giulio2002/gevm/spec"
	"github.com/Giulio2002/gevm/state"
	"github.com/Giulio2002/gevm/vm"
)

func newTestEVM(db state.Database, forkID gevmspec.ForkID, block host.BlockEnv, cfg host.CfgEnv) *host.Evm {
	evm := host.NewEvm(db, forkID, block, cfg)
	if runner := testRunner(forkID); runner != nil {
		evm.Set(runner)
	}
	return evm
}

func testRunner(forkID gevmspec.ForkID) vm.Runner {
	switch os.Getenv("GEVM_TEST_RUNNER") {
	case "", "default":
		return nil
	case "opcode":
		return vm.NewTracingRunner(&vm.Hooks{
			OnOpcode: func(uint64, byte, uint64, uint64, vm.OpContext, []byte, int, error) {},
		}, forkID)
	default:
		panic("unknown GEVM_TEST_RUNNER: " + os.Getenv("GEVM_TEST_RUNNER"))
	}
}
