package spec

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestFixtureReportCountsUnsupported(t *testing.T) {
	data, err := os.ReadFile("testdata/eest-v21.0.2/osaka-transfer.json")
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
		original := post["Osaka"][0]
		post["FutureFork"] = []map[string]json.RawMessage{original}
		rejected := make(map[string]json.RawMessage)
		for k, v := range original {
			rejected[k] = v
		}
		rejected["expectException"] = json.RawMessage(`"TransactionException.UNKNOWN"`)
		post["Osaka"] = append(post["Osaka"], rejected)
		unit["post"], _ = json.Marshal(post)
	}
	data, err = json.Marshal(suite)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err = os.WriteFile(filepath.Join(dir, "cases.json"), data, 0600); err != nil {
		t.Fatal(err)
	}
	r, err := RunFixtureReport(dir, RunnerConfig{})
	if err != nil {
		t.Fatal(err)
	}
	if r.Total != 3 || r.Passed != 1 || r.Failed != 0 || r.Unsupported != 2 || r.Skipped != 0 {
		t.Fatalf("incorrect report: %+v", r)
	}
	if _, err = RunFixtureReport(filepath.Join(dir, "missing"), RunnerConfig{}); err == nil {
		t.Fatal("missing directory must be an error")
	}
}
