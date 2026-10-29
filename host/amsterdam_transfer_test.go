package host

import (
	"bytes"
	"github.com/Giulio2002/gevm/spec"
	"github.com/Giulio2002/gevm/state"
	"github.com/Giulio2002/gevm/types"
	"testing"
)

func TestAmsterdamCallTransferLogs(t *testing.T) {
	for _, fork := range []spec.ForkID{spec.Osaka, spec.Amsterdam} {
		for _, tc := range []struct {
			name  string
			to    byte
			value uint64
			code  []byte
			logs  int
		}{
			{"transfer", 0x20, 7, nil, 1}, {"self", 1, 7, nil, 0}, {"zero", 0x20, 0, nil, 0},
			{"revert", 0x20, 7, []byte{0x5f, 0x5f, 0xfd}, 0}, {"halt", 0x20, 7, []byte{0xfe}, 0},
		} {
			t.Run(fork.String()+"/"+tc.name, func(t *testing.T) {
				db := newMockDB()
				db.accounts[addr(1)] = &state.AccountInfo{Balance: u(1_000_000_000), CodeHash: types.KeccakEmpty}
				if tc.to != 1 {
					db.accounts[addr(tc.to)] = &state.AccountInfo{Balance: u(1), Code: tc.code, CodeHash: types.Keccak256(tc.code)}
				}
				e := makeEvm(db, fork, BlockEnv{})
				defer e.ReleaseEvm()
				tx := simpleTx(addr(1), addr(tc.to), tc.value, 1_000_000, 1)
				r := e.Transact(&tx)
				want := tc.logs
				if fork == spec.Osaka {
					want = 0
				}
				if len(r.Logs) != want {
					t.Fatalf("logs=%d want %d: %+v", len(r.Logs), want, r)
				}
				if want == 0 {
					return
				}
				l := r.Logs[0]
				sys, _ := types.HexToAddress("0xfffffffffffffffffffffffffffffffffffffffe")
				a, b, v := addr(1), addr(tc.to), u(tc.value)
				sa, ra := a.ToU256(), b.ToU256()
				sender := sa.Bytes32()
				recipient := ra.Bytes32()
				amount := v.Bytes32()
				if l.Address != sys || l.NumTopics != 3 || l.Topics[0] != types.Keccak256([]byte("Transfer(address,address,uint256)")) || l.Topics[1] != types.B256(sender) || l.Topics[2] != types.B256(recipient) || !bytes.Equal(l.Data, amount[:]) {
					t.Fatalf("invalid transfer log: %+v", l)
				}
			})
		}
	}
}
