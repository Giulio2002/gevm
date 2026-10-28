package spec

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/Giulio2002/gevm/types"
)

// FixtureCaseReport is one post/fork/index case, including exclusions.
type FixtureCaseReport struct {
	File   string `json:"file"`
	Test   string `json:"test"`
	Fork   string `json:"fork"`
	Index  int    `json:"index"`
	Status string `json:"status"`
	Detail string `json:"detail,omitempty"`
}

// FixtureReport does not equate unsupported or filtered cases with passes.
// Passed cases have executed and matched both the full MPT root and logs hash.
type FixtureReport struct {
	Files       int                 `json:"files"`
	Total       int                 `json:"total"`
	Passed      int                 `json:"passed"`
	Failed      int                 `json:"failed"`
	Unsupported int                 `json:"unsupported"`
	Skipped     int                 `json:"skipped"`
	Cases       []FixtureCaseReport `json:"cases"`
}

// RunFixtureReport executes state fixtures with strict accounting. Transaction
// rejection fixtures are unsupported until exact exception matching is wired
// up: the legacy runner's "any validation error" test is not sufficient.
// Malformed/empty input is an error, never a successful zero-case run.
func RunFixtureReport(dir string, cfg RunnerConfig) (FixtureReport, error) {
	var report FixtureReport
	files, err := FindJSONTests(dir)
	if err != nil {
		return report, err
	}
	for _, path := range files {
		data, err := os.ReadFile(path)
		if err != nil {
			return report, err
		}
		var suite TestSuite
		if err = json.Unmarshal(data, &suite); err != nil {
			return report, fmt.Errorf("%s: %w", path, err)
		}
		if len(suite) == 0 {
			return report, fmt.Errorf("%s: empty state fixture", path)
		}
		report.Files++
		names := make([]string, 0, len(suite))
		for n := range suite {
			names = append(names, n)
		}
		sort.Strings(names)
		for _, name := range names {
			unit := suite[name]
			if unit == nil || len(unit.Post) == 0 {
				return report, fmt.Errorf("%s/%s: missing state-test post cases", path, name)
			}
			forks := make([]string, 0, len(unit.Post))
			for f := range unit.Post {
				forks = append(forks, f)
			}
			sort.Strings(forks)
			for _, fork := range forks {
				forkID, known := SpecNameToForkID(fork)
				for i, tc := range unit.Post[fork] {
					c := FixtureCaseReport{File: path, Test: name, Fork: fork, Index: i}
					switch {
					case skipTest(path, cfg) || (len(cfg.OnlySpecs) > 0 && !cfg.OnlySpecs[fork]):
						c.Status = "skipped"
						c.Detail = "excluded by configuration"
						report.Skipped++
					case !known:
						c.Status = "unsupported"
						c.Detail = "unknown fork"
						report.Unsupported++
					case tc.ExpectException != nil:
						c.Status = "unsupported"
						c.Detail = "exact transaction-exception matching not implemented: " + *tc.ExpectException
						report.Unsupported++
					case tc.Hash.V == types.B256Zero || tc.Logs.V == types.B256Zero:
						c.Status = "unsupported"
						c.Detail = "missing state root or logs hash"
						report.Unsupported++
					default:
						r := executeSingleTest(path, name, fork, forkID, unit, &tc, i)
						if r.Pass {
							c.Status = "passed"
							report.Passed++
						} else {
							c.Status = "failed"
							c.Detail = r.Detail
							report.Failed++
						}
					}
					report.Total++
					report.Cases = append(report.Cases, c)
				}
			}
		}
	}
	if report.Total == 0 {
		return report, fmt.Errorf("no state-test cases found in %s", dir)
	}
	return report, nil
}
