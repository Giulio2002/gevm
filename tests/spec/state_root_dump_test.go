package spec

import (
	"encoding/json"
	gevmspec "github.com/Giulio2002/gevm/spec"
	"github.com/Giulio2002/gevm/state"
	"os"
	"testing"
)

func TestOfficialStateDumpRoot(t *testing.T) {
	data, err := os.ReadFile("testdata/eest-v21.0.2/osaka-transfer.json")
	if err != nil {
		t.Fatal(err)
	}
	var suite TestSuite
	if err = json.Unmarshal(data, &suite); err != nil {
		t.Fatal(err)
	}
	for _, u := range suite {
		for _, cases := range u.Post {
			for _, tc := range cases {
				db := BuildMemDB(tc.State)
				j := state.NewJournal(db)
				root := stateRoot(db, j, gevmspec.Osaka)
				if root != tc.Hash.V {
					t.Fatalf("dump root %s want %s", root.Hex(), tc.Hash.V.Hex())
				}
			}
		}
	}
}
