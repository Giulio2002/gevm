package vm

// Comparison and bitwise opcode handlers.

import "github.com/holiman/uint256"

// opLt — BinaryOp body.
func opLt(interp *Interpreter) {
	s := interp.Stack
	if s.data[s.top].Lt(&s.data[s.top-1]) {
		s.data[s.top-1] = uint256.Int{1, 0, 0, 0}
	} else {
		s.data[s.top-1] = uint256.Int{}
	}
}

// opGt — BinaryOp body.
func opGt(interp *Interpreter) {
	s := interp.Stack
	if s.data[s.top].Gt(&s.data[s.top-1]) {
		s.data[s.top-1] = uint256.Int{1, 0, 0, 0}
	} else {
		s.data[s.top-1] = uint256.Int{}
	}
}

// opSlt — BinaryOp body.
func opSlt(interp *Interpreter) {
	s := interp.Stack
	a := &s.data[s.top]
	b := &s.data[s.top-1]
	aNeg := a[3] >> 63
	bNeg := b[3] >> 63
	var lt bool
	if aNeg != bNeg {
		lt = aNeg > bNeg
	} else {
		lt = a.Lt(b)
	}
	if lt {
		s.data[s.top-1] = uint256.Int{1, 0, 0, 0}
	} else {
		s.data[s.top-1] = uint256.Int{}
	}
}

// opSgt — BinaryOp body.
func opSgt(interp *Interpreter) {
	s := interp.Stack
	a := &s.data[s.top]
	b := &s.data[s.top-1]
	aNeg := a[3] >> 63
	bNeg := b[3] >> 63
	var gt bool
	if aNeg != bNeg {
		gt = bNeg > aNeg
	} else {
		gt = a.Gt(b)
	}
	if gt {
		s.data[s.top-1] = uint256.Int{1, 0, 0, 0}
	} else {
		s.data[s.top-1] = uint256.Int{}
	}
}

// opEq — BinaryOp body.
func opEq(interp *Interpreter) {
	s := interp.Stack
	if s.data[s.top].Eq(&s.data[s.top-1]) {
		s.data[s.top-1] = uint256.Int{1, 0, 0, 0}
	} else {
		s.data[s.top-1] = uint256.Int{}
	}
}

// opIszero — UnaryOp body.
func opIszero(interp *Interpreter) {
	s := interp.Stack
	if s.data[s.top-1].IsZero() {
		s.data[s.top-1] = uint256.Int{1, 0, 0, 0}
	} else {
		s.data[s.top-1] = uint256.Int{}
	}
}

// opAnd — BinaryOp body.
func opAnd(interp *Interpreter) {
	s := interp.Stack
	s.data[s.top-1].And(&s.data[s.top], &s.data[s.top-1])
}

// opOr — BinaryOp body.
func opOr(interp *Interpreter) {
	s := interp.Stack
	s.data[s.top-1].Or(&s.data[s.top], &s.data[s.top-1])
}

// opXor — BinaryOp body.
func opXor(interp *Interpreter) {
	s := interp.Stack
	s.data[s.top-1].Xor(&s.data[s.top], &s.data[s.top-1])
}

// opNot — UnaryOp body.
func opNot(interp *Interpreter) {
	s := interp.Stack
	s.data[s.top-1].Not(&s.data[s.top-1])
}

// opByte — BinaryOp body.
func opByte(interp *Interpreter) {
	s := interp.Stack
	a := s.data[s.top]
	top := &s.data[s.top-1]
	idx, overflow := a.Uint64WithOverflow()
	if idx < 32 {
		index := *uint256.NewInt(idx)
		top.Byte(&index)
	} else if overflow {
		*top = uint256.Int{}
	} else {
		*top = uint256.Int{}
	}
}

// opShl — BinaryOp body. Fork gate (Constantinople) handled by generator.
func opShl(interp *Interpreter) {
	s := interp.Stack
	shift := s.data[s.top]
	top := &s.data[s.top-1]
	sa, overflow := shift.Uint64WithOverflow()
	if !overflow && sa < 256 {
		top.Lsh(top, uint(sa))
	} else {
		*top = uint256.Int{}
	}
}

// opShr — BinaryOp body. Fork gate (Constantinople) handled by generator.
func opShr(interp *Interpreter) {
	s := interp.Stack
	shift := s.data[s.top]
	top := &s.data[s.top-1]
	sa, overflow := shift.Uint64WithOverflow()
	if !overflow && sa < 256 {
		top.Rsh(top, uint(sa))
	} else {
		*top = uint256.Int{}
	}
}

// opSar — BinaryOp body. Fork gate (Constantinople) handled by generator.
func opSar(interp *Interpreter) {
	s := interp.Stack
	shift := s.data[s.top]
	top := &s.data[s.top-1]
	sa, overflow := shift.Uint64WithOverflow()
	if !overflow && sa < 256 {
		top.SRsh(top, uint(sa))
	} else if top[3]&(1<<63) != 0 {
		*top = uint256.Int{^uint64(0), ^uint64(0), ^uint64(0), ^uint64(0)}
	} else {
		*top = uint256.Int{}
	}
}

// opClz — UnaryOp body. Fork gate (Osaka) handled by generator.
func opClz(interp *Interpreter) {
	s := interp.Stack
	top := &s.data[s.top-1]
	*top = *uint256.NewInt(uint64(256 - top.BitLen()))
}
