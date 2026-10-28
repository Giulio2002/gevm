package host

import (
	"fmt"
	"testing"

	"github.com/Giulio2002/gevm/spec"
	"github.com/Giulio2002/gevm/state"
	"github.com/Giulio2002/gevm/types"
	"github.com/Giulio2002/gevm/vm"
	"github.com/holiman/uint256"
)

func TestSelfdestructRollbackConservesBalance(t *testing.T) {
	for _, fork := range []spec.ForkID{spec.Osaka, spec.Amsterdam} {
		for _, mode := range []string{"default", "opcode"} {
			for _, outside := range []bool{false, true} {
				for _, halt := range []bool{false, true} {
					for _, repeats := range []int{1, 2} {
						t.Run(fmt.Sprintf("%s/%s/createdOutside=%t/exceptional=%t/repeats=%d", fork, mode, outside, halt, repeats), func(t *testing.T) {
							sender, outer, ancestor := addr(1), addr(0x20), addr(0x21)
							creator := ancestor
							if outside {
								creator = outer
							}
							child := types.CreateAddress(creator, 1)
							// Initcode returns ADDRESS; SELFDESTRUCT as the runtime.
							initcode := []byte{0x61, 0x30, 0xff, 0x5f, 0x52, 0x60, 0x02, 0x60, 0x1e, 0xf3}
							create := append([]byte{byte(0x5f + len(initcode))}, initcode...)
							create = append(create, 0x5f, 0x52, 0x60, byte(len(initcode)), 0x60, byte(32-len(initcode)), 0x60, 42, 0xf0, 0x50)
							call := func(to types.Address) []byte {
								// CALL with no value/data and a bounded execution grant.
								code := []byte{0x5f, 0x5f, 0x5f, 0x5f, 0x5f, 0x73}
								code = append(code, to[:]...)
								return append(code, 0x62, 0x0f, 0x42, 0x40, 0xf1)
							}
							var outerCode, ancestorCode []byte
							if outside {
								outerCode = append(outerCode, create...)
							} else {
								ancestorCode = append(ancestorCode, create...)
							}
							for i := 0; i < repeats; i++ {
								ancestorCode = append(ancestorCode, call(child)...)
								ancestorCode = append(ancestorCode, 0x50)
							}
							if halt {
								ancestorCode = append(ancestorCode, 0xfe) // exceptional INVALID halt
							} else {
								ancestorCode = append(ancestorCode, 0x5f, 0x5f, 0xfd) // REVERT
							}
							outerCode = append(outerCode, call(ancestor)...)
							// Return the failed ancestor CALL's success flag.
							outerCode = append(outerCode, 0x5f, 0x52, 0x60, 32, 0x5f, 0xf3)
							db := newMockDB()
							db.accounts[sender] = &state.AccountInfo{Balance: u(1000), CodeHash: types.KeccakEmpty}
							db.accounts[outer] = &state.AccountInfo{Balance: u(1000), Nonce: 1, Code: outerCode, CodeHash: types.Keccak256(outerCode)}
							db.accounts[ancestor] = &state.AccountInfo{Balance: u(1000), Nonce: 1, Code: ancestorCode, CodeHash: types.Keccak256(ancestorCode)}
							e := makeEvm(db, fork, BlockEnv{})
							defer e.ReleaseEvm()
							destructions := 0
							if mode == "opcode" {
								e.Set(vm.NewTracingRunner(&vm.Hooks{OnOpcode: func(_ uint64, op byte, _ uint64, _ uint64, _ vm.OpContext, _ []byte, _ int, _ error) {
									if op == 0xff {
										destructions++
									}
								}}, fork))
							}
							// Zero gas price/base fee isolates value conservation from fees.
							tx := simpleTx(sender, outer, 0, 10_000_000, 0)
							r := e.Transact(&tx)
							if !r.IsSuccess() || len(r.Output) != 32 || new(uint256.Int).SetBytes(r.Output).Sign() != 0 {
								t.Fatalf("outer must succeed after failed ancestor CALL: %+v", r)
							}
							if mode == "opcode" && destructions != repeats {
								t.Fatalf("executed SELFDESTRUCTs=%d want %d", destructions, repeats)
							}
							e.Journal.CommitTx()
							var total uint256.Int
							for _, acc := range e.Journal.State {
								total.Add(&total, &acc.Info.Balance)
							}
							if total != u(3000) {
								t.Errorf("total balance after rollback=%s want 3000", total.Dec())
							}
							acc := e.Journal.State[child]
							wantChild, wantCreator := u(0), u(1000)
							if outside {
								wantChild, wantCreator = u(42), u(958)
							}
							if acc == nil || acc.Info.Balance != wantChild || e.Journal.State[creator].Info.Balance != wantCreator {
								t.Fatalf("rollback must restore child/creator balances; child=%+v creator=%s", acc, e.Journal.State[creator].Info.Balance.Dec())
							}
							if outside && (acc.Info.Nonce != 1 || acc.Info.CodeHash != types.Keccak256([]byte{0x30, 0xff}) || acc.IsSelfdestructed()) {
								t.Fatal("reverted destruction deleted surviving child")
							}
							if !outside && (acc.Info.Nonce != 0 || len(acc.Info.Code) != 0 || acc.IsCreatedLocally()) {
								t.Fatal("reverted creation left child state")
							}
						})
					}
				}
			}
		}
	}
}
