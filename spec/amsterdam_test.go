package spec

import "testing"

func TestAmsterdamGasParamsAvailable(t *testing.T) {
	g := NewGasParams(Amsterdam)
	if g == nil || g != NewGasParams(Amsterdam) {
		t.Fatal("Amsterdam must have cached gas parameters")
	}
	if g.Get(GasIdExpByteGas) != 50 {
		t.Fatal("Amsterdam must retain EXP pricing")
	}
}
