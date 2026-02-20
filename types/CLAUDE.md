# primitives

Bit-exact EVM primitive types: `Uint256` ([4]uint64 little-endian limbs), `Address` ([20]byte), `B256` ([32]byte), `Bytes` ([]byte alias).

## Key types

- **Uint256** — 256-bit unsigned integer. Value receiver, no heap. Wrapping/checked/overflowing arithmetic matching Rust semantics.
- **Address** — 20-byte Ethereum address with `ToU256()` / `FromU256()` conversion.
- **B256** — 32-byte hash with `ToU256()` conversion.

## Rules

- `big.Int` is forbidden in hot paths.
- Shifts ≥ 256 return zero (wrapping semantics).
- Signed ops (SDIV, SMOD, SAR, SIGNEXTEND) use two's complement on the limb array directly.
- `Keccak256(data)` and `KeccakEmpty` live here (delegate to `internal/keccak`).

## Dependencies

`internal/keccak`
