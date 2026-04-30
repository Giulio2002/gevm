// Environment opcode handlers: ADDRESS through EXTCODEHASH.
package vm

import (
	"github.com/Giulio2002/gevm/spec"
	"github.com/Giulio2002/gevm/types"
	"github.com/holiman/uint256"
)

// opAddress — PushVal body.
func opAddress(interp *Interpreter) {
	s := interp.Stack
	s.data[s.top] = interp.Input.TargetAddress.ToU256()
	s.top++
}

// opBalance — Custom flush handler (needs Host).
func opBalance(interp *Interpreter, host Host) {
	s := interp.Stack
	if s.top == 0 {
		interp.HaltUnderflow()
		return
	}
	top := &s.data[s.top-1]
	addr := types.Address(top.Bytes20())
	balance, isCold := host.Balance(addr)
	if interp.RuntimeFlag.ForkID.IsEnabledIn(spec.Berlin) && isCold {
		cost := interp.GasParams.ColdAccountAdditionalCost()
		if !interp.Gas.RecordCost(cost) {
			interp.HaltOOG()
			return
		}
	}
	*top = balance
}

// opOrigin — PushVal body (needs Host).
func opOrigin(interp *Interpreter, host Host) {
	s := interp.Stack
	addr := host.Caller()
	s.data[s.top] = addr.ToU256()
	s.top++
}

// opCaller — PushVal body.
func opCaller(interp *Interpreter) {
	s := interp.Stack
	s.data[s.top] = interp.Input.CallerAddress.ToU256()
	s.top++
}

// opCallvalue — PushVal body.
func opCallvalue(interp *Interpreter) {
	s := interp.Stack
	s.data[s.top] = interp.Input.CallValue
	s.top++
}

// opCalldataload — UnaryOp body.
func opCalldataload(interp *Interpreter) {
	s := interp.Stack
	top := &s.data[s.top-1]
	offset, overflow := top.Uint64WithOverflow()
	if overflow {
		offset = ^uint64(0)
	}
	input := interp.Input.Input
	var word [32]byte
	if offset < uint64(len(input)) {
		src := input[offset:]
		if len(src) >= 32 {
			copy(word[:], src[:32])
		} else {
			copy(word[:], src)
		}
	}
	*top = *new(uint256.Int).SetBytes32((word)[:])
}

// opCalldatasize — PushVal body.
func opCalldatasize(interp *Interpreter) {
	s := interp.Stack
	s.data[s.top] = *uint256.NewInt(uint64(len(interp.Input.Input)))
	s.top++
}

// opCalldatacopy — Custom flush handler.
func opCalldatacopy(interp *Interpreter) {
	s := interp.Stack
	if s.top < 3 {
		interp.HaltUnderflow()
		return
	}
	s.top -= 3
	memOffsetVal := s.data[s.top+2]
	dataOffsetVal := s.data[s.top+1]
	lenVal := s.data[s.top]
	length, ok := interp.asUsizeOrFail(lenVal)
	if !ok {
		return
	}
	cost := interp.GasParams.CopyCost(uint64(length))
	if !interp.Gas.RecordCost(cost) {
		interp.HaltOOG()
		return
	}
	if length == 0 {
		return
	}
	memOffset, ok := interp.asUsizeOrFail(memOffsetVal)
	if !ok {
		return
	}
	if !interp.ResizeMemory(memOffset, length) {
		return
	}
	dataOffsetSat, overflow := dataOffsetVal.Uint64WithOverflow()
	if overflow {
		dataOffsetSat = ^uint64(0)
	}
	dataOffset := int(dataOffsetSat)
	if dataOffsetSat > uint64(maxInt) {
		dataOffset = maxInt
	}
	interp.Memory.SetData(memOffset, dataOffset, length, interp.Input.Input)
}

// opCodesize — PushVal body.
func opCodesize(interp *Interpreter) {
	s := interp.Stack
	s.data[s.top] = *uint256.NewInt(uint64(interp.Bytecode.originalLen))
	s.top++
}

// opCodecopy — Custom flush handler.
func opCodecopy(interp *Interpreter) {
	s := interp.Stack
	if s.top < 3 {
		interp.HaltUnderflow()
		return
	}
	s.top -= 3
	memOffsetVal := s.data[s.top+2]
	codeOffsetVal := s.data[s.top+1]
	lenVal := s.data[s.top]
	length, ok := interp.asUsizeOrFail(lenVal)
	if !ok {
		return
	}
	cost := interp.GasParams.CopyCost(uint64(length))
	if !interp.Gas.RecordCost(cost) {
		interp.HaltOOG()
		return
	}
	if length == 0 {
		return
	}
	memOffset, ok := interp.asUsizeOrFail(memOffsetVal)
	if !ok {
		return
	}
	if !interp.ResizeMemory(memOffset, length) {
		return
	}
	codeOffsetSat, overflow := codeOffsetVal.Uint64WithOverflow()
	if overflow {
		codeOffsetSat = ^uint64(0)
	}
	codeOffset := int(codeOffsetSat)
	if codeOffsetSat > uint64(maxInt) {
		codeOffset = maxInt
	}
	interp.Memory.SetData(memOffset, codeOffset, length, interp.Bytecode.code[:interp.Bytecode.originalLen])
}

// opGasprice — PushVal body (needs Host).
func opGasprice(interp *Interpreter, host Host) {
	s := interp.Stack
	s.data[s.top] = host.EffectiveGasPrice()
	s.top++
}

// opExtcodesize — Custom flush handler (needs Host).
func opExtcodesize(interp *Interpreter, host Host) {
	s := interp.Stack
	if s.top == 0 {
		interp.HaltUnderflow()
		return
	}
	top := &s.data[s.top-1]
	addr := types.Address(top.Bytes20())
	size, isCold := host.CodeSize(addr)
	if interp.RuntimeFlag.ForkID.IsEnabledIn(spec.Berlin) && isCold {
		cost := interp.GasParams.ColdAccountAdditionalCost()
		if !interp.Gas.RecordCost(cost) {
			interp.HaltOOG()
			return
		}
	}
	*top = *uint256.NewInt(uint64(size))
}

// opExtcodecopy — Custom flush handler (needs Host).
func opExtcodecopy(interp *Interpreter, host Host) {
	s := interp.Stack
	if s.top < 4 {
		interp.HaltUnderflow()
		return
	}
	s.top -= 4
	addrVal := s.data[s.top+3]
	memOffsetVal := s.data[s.top+2]
	codeOffsetVal := s.data[s.top+1]
	lenVal := s.data[s.top]
	addr := types.Address(addrVal.Bytes20())
	length, ok := interp.asUsizeOrFail(lenVal)
	if !ok {
		return
	}
	code, isCold := host.Code(addr)
	if interp.RuntimeFlag.ForkID.IsEnabledIn(spec.Berlin) && isCold {
		cost := interp.GasParams.ColdAccountAdditionalCost()
		if !interp.Gas.RecordCost(cost) {
			interp.HaltOOG()
			return
		}
	}
	cost := interp.GasParams.ExtcodecopyGas(uint64(length))
	if !interp.Gas.RecordCost(cost) {
		interp.HaltOOG()
		return
	}
	if length == 0 {
		return
	}
	memOffset, ok := interp.asUsizeOrFail(memOffsetVal)
	if !ok {
		return
	}
	if !interp.ResizeMemory(memOffset, length) {
		return
	}
	codeOffsetSat, overflow := codeOffsetVal.Uint64WithOverflow()
	if overflow {
		codeOffsetSat = ^uint64(0)
	}
	codeOffset := int(codeOffsetSat)
	if codeOffsetSat > uint64(maxInt) {
		codeOffset = maxInt
	}
	interp.Memory.SetData(memOffset, codeOffset, length, code)
}

// opReturndatasize — PushVal body.
func opReturndatasize(interp *Interpreter) {
	s := interp.Stack
	s.data[s.top] = *uint256.NewInt(uint64(len(interp.ReturnData)))
	s.top++
}

// opReturndatacopy — Custom flush handler.
func opReturndatacopy(interp *Interpreter) {
	s := interp.Stack
	if s.top < 3 {
		interp.HaltUnderflow()
		return
	}
	s.top -= 3
	memOffsetVal := s.data[s.top+2]
	dataOffsetVal := s.data[s.top+1]
	lenVal := s.data[s.top]
	length, ok := interp.asUsizeOrFail(lenVal)
	if !ok {
		return
	}
	dataOffset, ok := interp.asUsizeOrFail(dataOffsetVal)
	if !ok {
		return
	}
	end := dataOffset + length
	if end < dataOffset || end > len(interp.ReturnData) {
		interp.Halt(InstructionResultOutOfOffset)
		return
	}
	cost := interp.GasParams.CopyCost(uint64(length))
	if !interp.Gas.RecordCost(cost) {
		interp.HaltOOG()
		return
	}
	if length == 0 {
		return
	}
	memOffset, ok := interp.asUsizeOrFail(memOffsetVal)
	if !ok {
		return
	}
	if !interp.ResizeMemory(memOffset, length) {
		return
	}
	interp.Memory.Set(memOffset, interp.ReturnData[dataOffset:end])
}

// opExtcodehash — Custom flush handler (needs Host). Fork gate (Constantinople) checked by generator.
func opExtcodehash(interp *Interpreter, host Host) {
	s := interp.Stack
	if s.top == 0 {
		interp.HaltUnderflow()
		return
	}
	top := &s.data[s.top-1]
	addr := types.Address(top.Bytes20())
	hash, isCold := host.CodeHash(addr)
	if interp.RuntimeFlag.ForkID.IsEnabledIn(spec.Berlin) && isCold {
		cost := interp.GasParams.ColdAccountAdditionalCost()
		if !interp.Gas.RecordCost(cost) {
			interp.HaltOOG()
			return
		}
	}
	*top = hash.ToU256()
}
