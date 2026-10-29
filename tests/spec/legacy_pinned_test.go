package spec

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestLegacyPinnedFixtures(t *testing.T) {
	for _, runner := range []string{"default", "opcode"} {
		t.Run(runner, func(t *testing.T) {
			t.Setenv("GEVM_TEST_RUNNER", runner)
			report, err := RunFixtureReport("testdata/eest-v5.4.0", RunnerConfig{})
			if err != nil {
				t.Fatal(err)
			}
			if report.Files != 5 || report.Total != 32 || report.Passed != 32 || report.Failed != 0 || report.Unsupported != 0 || report.Skipped != 0 {
				t.Fatalf("legacy fixtures: %+v", report)
			}
		})
	}
}

func TestLegacyPreForkRejections(t *testing.T) {
	for _, runner := range []string{"default", "opcode"} {
		t.Run(runner, func(t *testing.T) {
			t.Setenv("GEVM_TEST_RUNNER", runner)
			rejected := 0
			for _, file := range []string{"test_eip2930_tx_validity.json", "test_eip1559_tx_validity.json", "test_blob_type_tx_pre_fork.json", "test_set_code_type_tx_pre_fork.json"} {
				data, err := os.ReadFile(filepath.Join("testdata/eest-v5.4.0", file))
				if err != nil {
					t.Fatal(err)
				}
				var suite TestSuite
				if err := json.Unmarshal(data, &suite); err != nil {
					t.Fatal(err)
				}
				for name, unit := range suite {
					for forkName, cases := range unit.Post {
						fork, ok := SpecNameToForkID(forkName)
						if !ok {
							t.Fatal(forkName)
						}
						for _, tc := range cases {
							if tc.ExpectException == nil {
								result := executeSingleTest(file, name, forkName, fork, unit, &tc, 0)
								if !result.Pass {
									t.Errorf("%s: %s", name, result.Detail)
								}
								continue
							}
							rejected++
							if !knownExceptions(*tc.ExpectException) {
								t.Errorf("unsupported: %s", *tc.ExpectException)
							}
							if err := executeExceptionCase(unit, &tc, fork); err != nil {
								t.Errorf("%s: %v", name, err)
							}
							wrong := tc
							unrelated := "TransactionException.TYPE_1_TX_PRE_FORK"
							if *tc.ExpectException == unrelated {
								unrelated = "TransactionException.TYPE_2_TX_PRE_FORK"
							}
							wrong.ExpectException = &unrelated
							if executeExceptionCase(unit, &wrong, fork) == nil {
								t.Errorf("%s accepted unrelated exception", name)
							}
							wrong = tc
							wrong.Hash.V[0] ^= 1
							if executeExceptionCase(unit, &wrong, fork) == nil {
								t.Errorf("%s accepted wrong root", name)
							}
							wrong = tc
							wrong.Logs.V[0] ^= 1
							if executeExceptionCase(unit, &wrong, fork) == nil {
								t.Errorf("%s accepted wrong logs hash", name)
							}
						}
					}
				}
			}
			if rejected != 15 {
				t.Fatalf("checked %d rejections, want 15", rejected)
			}
		})
	}
}
