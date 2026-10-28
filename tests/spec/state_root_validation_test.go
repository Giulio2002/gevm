package spec

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestStateFixtureRejectsWrongRoot(t *testing.T) {
	p := "testdata/eest-v21.0.2/osaka-transfer.json"
	rs, err := RunTestFile(p, DefaultConfig())
	if err != nil || len(rs) == 0 {
		t.Fatalf("fixture: %v", err)
	}
	for _, r := range rs {
		if !r.Pass {
			t.Fatalf("official fixture failed: %+v", r)
		}
	}
	data, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	var suite map[string]map[string]json.RawMessage
	if err = json.Unmarshal(data, &suite); err != nil {
		t.Fatal(err)
	}
	for _, unit := range suite {
		var post map[string][]map[string]json.RawMessage
		if err = json.Unmarshal(unit["post"], &post); err != nil {
			t.Fatal(err)
		}
		for _, cases := range post {
			for _, c := range cases {
				c["hash"] = json.RawMessage(`"0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"`)
			}
		}
		unit["post"], _ = json.Marshal(post)
	}
	data, err = json.Marshal(suite)
	if err != nil {
		t.Fatal(err)
	}
	p = filepath.Join(t.TempDir(), "wrong-root.json")
	if err = os.WriteFile(p, data, 0600); err != nil {
		t.Fatal(err)
	}
	rs, err = RunTestFile(p, DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range rs {
		if r.Pass || r.Error.Kind != TestErrStateRootMismatch {
			t.Fatalf("wrong root accepted: %+v", r)
		}
	}
}
