package spec

import (
	"path/filepath"
	"testing"
)

func TestRunTestDirMissingIsFailure(t *testing.T) {
	passed, failed, failures := RunTestDir(filepath.Join(t.TempDir(), "missing"), DefaultConfig())
	if passed != 0 || failed != 1 || len(failures) != 1 || failures[0].Detail == "" {
		t.Fatalf("missing path: passed=%d failed=%d failures=%+v", passed, failed, failures)
	}
}
