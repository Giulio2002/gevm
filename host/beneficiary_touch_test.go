package host

import (
	"testing"

	"github.com/Giulio2002/gevm/spec"
	"github.com/Giulio2002/gevm/state"
	"github.com/Giulio2002/gevm/types"
)

// Before EIP-161 even a zero fee creates an absent beneficiary account.
func TestZeroFeeCreatesBeneficiaryBeforeEIP161(t *testing.T) {
	for _, fork := range []spec.ForkID{spec.Frontier, spec.Homestead, spec.Tangerine, spec.SpuriousDragon, spec.London, spec.Amsterdam} {
		t.Run(fork.String(), func(t *testing.T) {
			db := newMockDB()
			db.accounts[addr(1)] = &state.AccountInfo{Balance: u(1000000), CodeHash: types.KeccakEmpty}
			evm := makeEvm(db, fork, BlockEnv{Beneficiary: addr(0xcb)})
			tx := simpleTx(addr(1), addr(0x20), 0, 21000, 0)
			result := evm.Transact(&tx)
			if result.Kind != ResultSuccess {
				t.Fatalf("transaction: %+v", result)
			}
			beneficiary, err := evm.Journal.LoadAccount(addr(0xcb))
			if err != nil {
				t.Fatal(err)
			}
			wantTouched := !fork.IsEnabledIn(spec.SpuriousDragon)
			if beneficiary.Data.IsTouched() != wantTouched {
				t.Fatalf("beneficiary touched = %v, want %v", beneficiary.Data.IsTouched(), wantTouched)
			}
			if !beneficiary.Data.Info.Balance.IsZero() {
				t.Fatal("zero fee increased balance")
			}
		})
	}
}
