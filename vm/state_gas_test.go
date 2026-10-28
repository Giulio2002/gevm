package vm

import "testing"

func TestStateGasCrossFrameRefund(t *testing.T) {
	parent := NewGas(1000)
	parent.SetStateGas(100)
	parent.RecordStateCost(300) // 200 execution gas spilled
	child := NewGas(100)
	child.SetStateGas(parent.DrainStateGas())
	child.RefundStateGas(300) // child did not incur the parent's spill
	parent.RecordCost(100)
	parent.MergeStateGas(child, true)
	if parent.Remaining() != 1000 || parent.StateGasRemaining() != 100 {
		t.Fatalf("cross-frame repayment: %+v", parent)
	}
	parent.RestoreStateGas()
	if parent.Remaining() != 1000 || parent.StateGasRemaining() != 100 {
		t.Fatalf("rollback after repayment: %+v", parent)
	}
}

func TestStateGasCommittedRollback(t *testing.T) {
	g := NewGas(1000)
	g.SetStateGas(100)
	g.RecordStateCost(300)
	g.CommitStateGas()
	g.RecordStateCost(50)
	g.RecordCost(10)
	g.RestoreStateGas()
	if g.Remaining() != 790 || g.StateGasRemaining() != 0 {
		t.Fatalf("committed charge refilled: %+v", g)
	}
	g.RestoreStateGasToEntry(100)
	if g.Remaining() != 990 || g.StateGasRemaining() != 100 {
		t.Fatalf("preparation rollback: %+v", g)
	}
}

func TestStateGasReservoirRollback(t *testing.T) {
	g := NewGas(1000)
	g.SetStateGas(200)
	if !g.RecordStateCost(350) {
		t.Fatal("charge failed")
	}
	if g.Remaining() != 850 || g.StateGasRemaining() != 0 {
		t.Fatalf("wrong pools: %+v", g)
	}
	if g.RecordStateCost(851) {
		t.Fatal("unaffordable state charge accepted")
	}
	g.RecordCost(100)
	g.RestoreStateGas()
	if g.Remaining() != 900 || g.StateGasRemaining() != 200 {
		t.Fatalf("rollback: %+v", g)
	}
}
