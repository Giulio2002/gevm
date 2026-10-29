package vm

// DrainStateGas forwards the entire reservoir; EIP-150 applies only to execution gas.
func (g *Gas) DrainStateGas() uint64 { n := g.stateRemaining; g.stateRemaining = 0; return n }

// RefundStateGas credits spill first (LIFO), then the reservoir.
func (g *Gas) RefundStateGas(amount uint64) {
	spill := min(amount, g.stateSpilled)
	g.remaining += spill
	g.stateSpilled -= spill
	g.stateRemaining += amount - spill
}

// MergeStateGas absorbs an already-settled child, including execution gas.
// A successful child's cross-frame refunds repay the parent's outstanding spill.
func (g *Gas) MergeStateGas(child Gas, success bool) {
	g.remaining += child.remaining
	g.stateRemaining += child.stateRemaining
	g.stateSpilled += child.stateSpilled
	g.refunded += child.refunded
	if success {
		repay := min(g.stateRemaining, g.stateSpilled)
		g.remaining += repay
		g.stateRemaining -= repay
		g.stateSpilled -= repay
	}
}

func (g *Gas) CommitStateGas() {
	g.stateBaseline = g.stateRemaining
	g.stateCommittedSpill += g.stateSpilled
	g.stateSpilled = 0
}

func (g *Gas) RestoreStateGasToEntry(grant uint64) {
	g.RestoreStateGas()
	g.remaining += g.stateCommittedSpill
	g.stateCommittedSpill = 0
	g.stateRemaining = grant
	g.stateBaseline = grant
}

// SetStateGas initializes a frame's EIP-8037 reservoir and rollback baseline.
func (g *Gas) SetStateGas(amount uint64) { g.stateRemaining = amount; g.stateBaseline = amount }
func (g *Gas) StateGasRemaining() uint64 { return g.stateRemaining }

// NetStateGasUsed measures the transaction's state dimension, including spill.
// Refunds from prior frames can make the un-clamped result negative.
func (g *Gas) NetStateGasUsed(grant uint64) int64 {
	return int64(grant) - int64(g.stateRemaining) + int64(g.stateSpilled) + int64(g.stateCommittedSpill)
}

// RecordStateCost draws from the reservoir before execution gas. A failed
// charge is atomic; a later frame rollback must not credit unpaid charges.
func (g *Gas) RecordStateCost(amount uint64) bool {
	if amount <= g.stateRemaining {
		g.stateRemaining -= amount
		return true
	}
	spill := amount - g.stateRemaining
	if spill > g.remaining {
		return false
	}
	g.remaining -= spill
	g.stateRemaining = 0
	g.stateSpilled += spill
	return true
}

// RestoreStateGas rolls back uncommitted state charges, not execution charges.
// On exceptional halt the caller must burn execution gas AFTER this operation.
func (g *Gas) RestoreStateGas() {
	g.remaining += g.stateSpilled
	g.stateSpilled = 0
	g.stateRemaining = g.stateBaseline
	g.refunded = 0
}
