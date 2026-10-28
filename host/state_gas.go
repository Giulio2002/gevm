package host

import "github.com/Giulio2002/gevm/vm"

func newFrameGas(execution, reservoir uint64) vm.Gas {
	gas := vm.NewGas(execution)
	gas.SetStateGas(reservoir)
	return gas
}

// State rollback precedes exceptional execution-gas forfeiture (EIP-8037).
func settleFrameGas(result *vm.InterpreterResult) {
	if !result.Result.IsOk() {
		result.Gas.RestoreStateGas()
	}
	if !result.Result.IsOkOrRevert() {
		result.Gas.SpendAll()
	}
}
