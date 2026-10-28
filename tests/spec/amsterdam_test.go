package spec

import (
	"encoding/json"
	gevmspec "github.com/Giulio2002/gevm/spec"
	"testing"
)

func TestAmsterdamSlotNumberEnvironment(t *testing.T) {
	var unit TestUnit
	if err := json.Unmarshal([]byte(`{"env":{"slotNumber":"0x123456789abcdef0"}}`), &unit); err != nil {
		t.Fatal(err)
	}
	got := BuildBlockEnv(&unit, gevmspec.Amsterdam).SlotNum
	if got.Uint64() != 0x123456789abcdef0 {
		t.Fatalf("slot number=%x", got.Uint64())
	}
}
