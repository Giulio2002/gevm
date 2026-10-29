package spec

import (
	"encoding/json"
	gevmspec "github.com/Giulio2002/gevm/spec"
	"os"
	"testing"
)

func TestAmsterdamPinnedFixtures(t *testing.T) {
	for _, runner := range []string{"default", "opcode"} {
		t.Run(runner, func(t *testing.T) {
			t.Setenv("GEVM_TEST_RUNNER", runner)
			report, err := RunFixtureReport("testdata/eest-v21.0.2/amsterdam", RunnerConfig{})
			if err != nil {
				t.Fatal(err)
			}
			if report.Files != 10 || report.Total != 18 || report.Passed != 18 || report.Failed != 0 || report.Unsupported != 0 || report.Skipped != 0 {
				t.Fatalf("pinned regression fixtures: %+v", report)
			}
		})
	}
}

func TestRejectionRequiresExactIdentityAndStateRoot(t *testing.T) {
	raw, err := os.ReadFile("testdata/eest-v21.0.2/amsterdam/tx_max_nonce.json")
	if err != nil {
		t.Fatal(err)
	}
	var suite TestSuite
	if err = json.Unmarshal(raw, &suite); err != nil {
		t.Fatal(err)
	}
	for _, unit := range suite {
		for _, tc := range unit.Post["Amsterdam"] {
			if err = executeExceptionCase(unit, &tc, gevmspec.Amsterdam); err != nil {
				t.Fatal(err)
			}
			wrong := tc
			s := "TransactionException.INSUFFICIENT_ACCOUNT_FUNDS"
			wrong.ExpectException = &s
			if executeExceptionCase(unit, &wrong, gevmspec.Amsterdam) == nil {
				t.Fatal("unrelated rejection accepted")
			}
			wrong = tc
			wrong.Hash.V[0] ^= 1
			if executeExceptionCase(unit, &wrong, gevmspec.Amsterdam) == nil {
				t.Fatal("wrong rejection root accepted")
			}
		}
	}
}
