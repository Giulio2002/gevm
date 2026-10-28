// State root MPT helpers. The complete fixture/journal overlay driver lives
// in state_root_overlay.go and is used by state fixture validation.
package spec

import (
	"github.com/Giulio2002/gevm/types"
	"github.com/holiman/uint256"
)

// storageRoot computes the MPT root of an account's storage trie.
func storageRoot(storage map[uint256.Int]uint256.Int) types.B256 {
	if len(storage) == 0 {
		return emptyTrieRoot
	}

	entries := make([]mptEntry, 0, len(storage))
	for slot, value := range storage {
		slotBytes := slot.Bytes32()
		slotHash := types.Keccak256(slotBytes[:])
		entries = append(entries, mptEntry{
			keyNibbles: keyToNibbles(slotHash),
			value:      RlpEncodeU256(value),
		})
	}

	return mptRoot(entries)
}

// rlpEncodeAccount encodes [nonce, balance, storageRoot, codeHash] as RLP.
func rlpEncodeAccount(nonce uint64, balance uint256.Int, sRoot, codeHash types.B256) []byte {
	return RlpEncodeList([][]byte{
		RlpEncodeUint64(nonce),
		RlpEncodeU256(balance),
		RlpEncodeBytes(sRoot[:]),
		RlpEncodeBytes(codeHash[:]),
	})
}
