package spec

import (
	gevmspec "github.com/Giulio2002/gevm/spec"
	"github.com/holiman/uint256"
	"testing"
)

func TestTransactionFixtureAmsterdamInitcodeLimit(t *testing.T) {
	for _, fork := range []gevmspec.ForkID{gevmspec.Osaka, gevmspec.Amsterdam} {
		for _, size := range []int{49152, 49153, 131072, 131073} {
			tx := DecodedTx{GasLimit: 10_000_000, Data: make([]byte, size), V: *uint256.NewInt(27), R: *uint256.NewInt(1), S: *uint256.NewInt(1)}
			err := validateTxForFork(&tx, fork)
			limit := 49152
			if fork == gevmspec.Amsterdam {
				limit = 131072
			}
			if (err == nil) != (size <= limit) {
				t.Errorf("%v/%d: %v", fork, size, err)
			}
		}
	}
}
