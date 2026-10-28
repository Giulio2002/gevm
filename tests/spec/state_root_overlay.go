package spec

import (
	gevmspec "github.com/Giulio2002/gevm/spec"
	"github.com/Giulio2002/gevm/state"
	"github.com/Giulio2002/gevm/types"
	"github.com/holiman/uint256"
)

// stateRoot merges the journal overlay with the complete fixture pre-state.
// Reading an account must not accidentally insert it into the trie. Storage
// zeroes remove leaves, and CREATE/selfdestruct clear pre-existing storage.
func stateRoot(db *MemDB, journal *state.Journal, fork gevmspec.ForkID) types.B256 {
	addresses := make(map[types.Address]struct{}, len(db.accounts)+len(journal.State))
	for a := range db.accounts {
		addresses[a] = struct{}{}
	}
	for a := range journal.State {
		addresses[a] = struct{}{}
	}
	entries := make([]mptEntry, 0, len(addresses))
	for addr := range addresses {
		pre := db.accounts[addr]
		acc := journal.State[addr]
		var info state.AccountInfo
		slots := make(map[uint256.Int]uint256.Int)
		if pre != nil {
			info = pre.info
			for k, v := range pre.storage {
				if !v.IsZero() {
					slots[k] = v
				}
			}
		}
		if acc != nil {
			if acc.IsLoadedAsNotExistingNotTouched() {
				continue
			}
			if acc.IsSelfdestructed() && (!fork.IsEnabledIn(gevmspec.Amsterdam) || acc.Info.Balance.IsZero()) {
				continue
			}
			if fork.IsEnabledIn(gevmspec.SpuriousDragon) && acc.IsTouched() && acc.IsEmpty() && !acc.IsPreserveEmpty() {
				continue
			}
			info = acc.Info
			if acc.IsCreated() || acc.IsStorageCleared() {
				clear(slots)
			}
			for k, v := range acc.Storage {
				if v.PresentValue.IsZero() {
					delete(slots, k)
				} else {
					slots[k] = v.PresentValue
				}
			}
			if acc.IsSelfdestructed() {
				// Amsterdam finalization clears the account except its balance.
				info = state.AccountInfo{Balance: info.Balance, CodeHash: types.KeccakEmpty}
				clear(slots)
			}
		}
		hash := info.CodeHash
		if hash == types.B256Zero {
			hash = types.KeccakEmpty
		}
		key := types.Keccak256(addr[:])
		entries = append(entries, mptEntry{keyNibbles: keyToNibbles(key), value: rlpEncodeAccount(info.Nonce, info.Balance, storageRoot(slots), hash)})
	}
	return mptRoot(entries)
}
