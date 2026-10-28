package vm

// decodeSingle decodes the EIP-8024 immediate for DUPN/SWAPN.
// The forbidden range keeps JUMPDEST and PUSH opcodes out of immediates,
// so legacy jump-destination analysis remains valid.
func decodeSingle(x int) (int, bool) {
	if x < 0 || x > 255 || (x > 90 && x < 128) {
		return 0, false
	}
	return (x + 145) % 256, true
}

// decodePair decodes the EIP-8024 EXCHANGE immediate into stack depths.
func decodePair(x int) (int, int, bool) {
	if x < 0 || x > 255 || (x > 81 && x < 128) {
		return 0, 0, false
	}
	k := x ^ 143
	q, r := k/16, k%16
	if q < r {
		return q + 1, r + 1, true
	}
	return r + 1, 29 - q, true
}
