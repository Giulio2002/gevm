package spec

import (
	"encoding/json"
	"github.com/Giulio2002/gevm/host"
	gevmspec "github.com/Giulio2002/gevm/spec"
	"os"
	"strings"
	"testing"
)

// TestEESTCaseDebug is an opt-in diagnostic for a single pinned fixture case.
func TestEESTCaseDebug(t *testing.T) {
	path := os.Getenv("GEVM_EEST_CASE")
	if path == "" {
		t.Skip("set GEVM_EEST_CASE")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var suite TestSuite
	if err = json.Unmarshal(raw, &suite); err != nil {
		t.Fatal(err)
	}
	for name, unit := range suite {
		if !strings.Contains(name, os.Getenv("GEVM_EEST_NAME")) {
			continue
		}
		for _, tc := range unit.Post["Amsterdam"] {
			caller, err := RecoverAddress(unit.Transaction.SecretKey.V[:])
			if err != nil {
				t.Fatal(err)
			}
			tx, err := BuildTransaction(unit, &tc, caller)
			if err != nil {
				t.Fatal(err)
			}
			db := BuildMemDB(unit.Pre)
			e := newTestEVM(db, gevmspec.Amsterdam, BuildBlockEnv(unit, gevmspec.Amsterdam), host.CfgEnv{ChainId: unit.ChainId()})
			result := e.Transact(&tx)
			t.Logf("%s result %+v", name, result)
			expected := tc.State
			if expected == nil {
				expected = tc.PostState
			}
			root := stateRoot(db, e.Journal, gevmspec.Amsterdam)
			t.Logf("root %s expected %s; dump error: %v", root.Hex(), tc.Hash.V.Hex(), validatePostState(e.Journal, expected))
			for a, ex := range expected {
				if acc := e.Journal.State[a.V]; acc != nil && acc.Info.Balance != ex.Balance.V {
					t.Logf("balance %s got %s want %s", a.V.Hex(), acc.Info.Balance.Dec(), ex.Balance.V.Dec())
				}
			}
		}
	}
}
