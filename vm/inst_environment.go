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
	s.data[s.top] = uint256.Int{uint64(len(interp.Input.Input)), 0, 0, 0}
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
	if lenVal[1]|lenVal[2]|lenVal[3] != 0 || lenVal[0] > uint64(maxInt) {
		interp.Halt(InstructionResultInvalidOperandOOG)
		return
	}
	length := int(lenVal[0])
	cost := interp.GasParams.CopyCost(uint64(length))
	if !interp.Gas.RecordCost(cost) {
		interp.HaltOOG()
		return
	}
	if length == 0 {
		return
	}
	if memOffsetVal[1]|memOffsetVal[2]|memOffsetVal[3] != 0 || memOffsetVal[0] > uint64(maxInt) {
		interp.Halt(InstructionResultInvalidOperandOOG)
		return
	}
	memOffset := int(memOffsetVal[0])
	if !interp.ResizeMemory(memOffset, length) {
		return
	}
	dataOffset := maxInt
	if dataOffsetVal[1]|dataOffsetVal[2]|dataOffsetVal[3] == 0 && dataOffsetVal[0] <= uint64(maxInt) {
		dataOffset = int(dataOffsetVal[0])
	}
	copyPaddedToMemory(interp.Memory, memOffset, dataOffset, length, interp.Input.Input)
}

func copyPaddedToMemory(memory *Memory, memoryOffset, dataOffset, length int, data []byte) {
	dst := (*memory.buffer)[memory.checkpoint+memoryOffset : memory.checkpoint+memoryOffset+length]
	if dataOffset < 0 || dataOffset >= len(data) {
		clear(dst)
		return
	}
	srcEnd := dataOffset + length
	if srcEnd > len(data) {
		srcEnd = len(data)
	}
	srcLen := srcEnd - dataOffset
	copy(dst[:srcLen], data[dataOffset:srcEnd])
	if srcLen < length {
		clear(dst[srcLen:])
	}
}

func saturatedU256ToInt(v uint256.Int) int {
	if v[1]|v[2]|v[3] != 0 || v[0] > uint64(maxInt) {
		return maxInt
	}
	return int(v[0])
}

// opCodesize — PushVal body.
func opCodesize(interp *Interpreter) {
	s := interp.Stack
	s.data[s.top] = uint256.Int{uint64(interp.Bytecode.originalLen), 0, 0, 0}
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
	if lenVal[1]|lenVal[2]|lenVal[3] != 0 || lenVal[0] > uint64(maxInt) {
		interp.Halt(InstructionResultInvalidOperandOOG)
		return
	}
	length := int(lenVal[0])
	cost := interp.GasParams.CopyCost(uint64(length))
	if !interp.Gas.RecordCost(cost) {
		interp.HaltOOG()
		return
	}
	if length == 0 {
		return
	}
	if memOffsetVal[1]|memOffsetVal[2]|memOffsetVal[3] != 0 || memOffsetVal[0] > uint64(maxInt) {
		interp.Halt(InstructionResultInvalidOperandOOG)
		return
	}
	memOffset := int(memOffsetVal[0])
	if !interp.ResizeMemory(memOffset, length) {
		return
	}
	codeOffset := saturatedU256ToInt(codeOffsetVal)
	copyPaddedToMemory(interp.Memory, memOffset, codeOffset, length, interp.Bytecode.code[:interp.Bytecode.originalLen])
}

// opGasprice — PushVal body (needs Host).
func opGasprice(interp *Interpreter, host Host) {
	s := interp.Stack
	s.data[s.top] = host.EffectiveGasPrice()
	s.top++
}

// opExtcodesize — Custom flush handler (needs Host).
func opExtcodesize(interp *Interpreter, host Host) {
	if interp.RuntimeFlag.ForkID.IsEnabledIn(spec.Amsterdam) && !interp.Gas.RecordCost(100) {
		interp.HaltOOG()
		return
	}
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
	*top = uint256.Int{uint64(size), 0, 0, 0}
}

// opExtcodecopy — Custom flush handler (needs Host).
func opExtcodecopy(interp *Interpreter, host Host) {
	if interp.RuntimeFlag.ForkID.IsEnabledIn(spec.Amsterdam) && !interp.Gas.RecordCost(100) {
		interp.HaltOOG()
		return
	}
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
	if lenVal[1]|lenVal[2]|lenVal[3] != 0 || lenVal[0] > uint64(maxInt) {
		interp.Halt(InstructionResultInvalidOperandOOG)
		return
	}
	length := int(lenVal[0])
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
	if memOffsetVal[1]|memOffsetVal[2]|memOffsetVal[3] != 0 || memOffsetVal[0] > uint64(maxInt) {
		interp.Halt(InstructionResultInvalidOperandOOG)
		return
	}
	memOffset := int(memOffsetVal[0])
	if !interp.ResizeMemory(memOffset, length) {
		return
	}
	codeOffset := saturatedU256ToInt(codeOffsetVal)
	copyPaddedToMemory(interp.Memory, memOffset, codeOffset, length, code)
}

// opReturndatasize — PushVal body.
func opReturndatasize(interp *Interpreter) {
	s := interp.Stack
	s.data[s.top] = uint256.Int{uint64(len(interp.ReturnData)), 0, 0, 0}
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
	if lenVal[1]|lenVal[2]|lenVal[3] != 0 || lenVal[0] > uint64(maxInt) {
		interp.Halt(InstructionResultInvalidOperandOOG)
		return
	}
	length := int(lenVal[0])
	if dataOffsetVal[1]|dataOffsetVal[2]|dataOffsetVal[3] != 0 || dataOffsetVal[0] > uint64(maxInt) {
		interp.Halt(InstructionResultInvalidOperandOOG)
		return
	}
	dataOffset := int(dataOffsetVal[0])
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
	if memOffsetVal[1]|memOffsetVal[2]|memOffsetVal[3] != 0 || memOffsetVal[0] > uint64(maxInt) {
		interp.Halt(InstructionResultInvalidOperandOOG)
		return
	}
	memOffset := int(memOffsetVal[0])
	if !interp.ResizeMemory(memOffset, length) {
		return
	}
	copy((*interp.Memory.buffer)[interp.Memory.checkpoint+memOffset:], interp.ReturnData[dataOffset:end])
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
