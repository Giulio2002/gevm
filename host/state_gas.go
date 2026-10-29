package host

import (
	"github.com/Giulio2002/gevm/types"
	"github.com/Giulio2002/gevm/vm"
)

// prepareCreateGas performs preflight and state charging before the child grant.
func (h *Handler) prepareCreateGas(interp *vm.Interpreter, depth int) (charged, ok bool) {
	input := &interp.ActionData.Create
	caller, err := h.Host.Journal.LoadAccount(input.Caller)
	if err != nil {
		interp.Halt(vm.InstructionResultFatalExternalError)
		return false, false
	}
	if depth <= CallStackLimit && caller.Data.Info.Nonce != ^uint64(0) && !caller.Data.Info.Balance.Lt(&input.Value) {
		target := types.CreateAddress(input.Caller, caller.Data.Info.Nonce)
		if input.Scheme.Kind == vm.CreateSchemeCreate2 {
			target = types.Create2Address(input.Caller, input.Scheme.Salt.Bytes32(), types.Keccak256(input.InitCode))
		}
		account, err := h.Host.Journal.LoadAccount(target)
		if err != nil {
			interp.Halt(vm.InstructionResultFatalExternalError)
			return false, false
		}
		if account.Data.Info.IsEmpty() {
			if !interp.Gas.RecordStateCost(120 * 1530) {
				interp.HaltOOG()
				return false, false
			}
			charged = true
		}
	}
	input.GasLimit = interp.Gas.Remaining() - interp.Gas.Remaining()/64
	interp.Gas.RecordCost(input.GasLimit)
	return charged, true
}

// prepareAmsterdamCall charges state-dependent top-frame preparation. The
// caller owns a journal checkpoint: any failure undoes ALL authorizations.
func (evm *Evm) prepareAmsterdamCall(tx *Transaction, gas *vm.Gas) bool {
	if len(tx.AuthorizationList) > 0 {
		if _, ok := evm.applyAuthList(tx, gas); !ok {
			return false
		}
		gas.CommitStateGas()
	}
	target, err := evm.Journal.LoadAccount(tx.To)
	if err != nil {
		return false
	}
	if !tx.Value.IsZero() && target.Data.Info.IsEmpty() && !gas.RecordStateCost(120*1530) {
		return false
	}
	code, err := evm.host.loadCode(tx.To, target.Data)
	if err != nil {
		return false
	}
	if delegate, ok := EIP7702Address(code); ok {
		load, err := evm.Journal.LoadAccount(delegate)
		if err != nil {
			return false
		}
		cost := uint64(100)
		if load.IsCold {
			cost = 3000
		}
		if !gas.RecordCost(cost) {
			return false
		}
	}
	return true
}

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
