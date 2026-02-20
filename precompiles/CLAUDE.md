# precompiles

Ethereum precompile contracts at fixed addresses. Fork-aware registry.

## Registry

| Fork | Addresses | Precompiles |
|------|-----------|-------------|
| Homestead | 0x01–0x04 | ECRECOVER, SHA256, RIPEMD160, IDENTITY |
| Byzantium | 0x05–0x08 | MODEXP, BN254_ADD, BN254_MUL, BN254_PAIRING |
| Istanbul | 0x09 | BLAKE2F |
| Cancun | 0x0A | KZG_POINT_EVALUATION (EIP-4844) |
| Prague | 0x0B–0x11 | BLS12-381 (7 operations, gnark-crypto) |
| Osaka | 0x0100 | P256VERIFY (secp256r1, Go stdlib) |

**P256VERIFY is Osaka-only, NOT Prague.** Including it in Prague warms address 0x0100, saving 2500 gas incorrectly.

## Key types

- **PrecompileSet** — Fork-specific set. `short [18]*Precompile` for 0x01–0x11 (O(1) lookup), `extended` map for 0x0100+.
- **PrecompileFn** — `func(input []byte, gasLimit uint64) PrecompileResult`
- **PrecompileResult** — Success (PrecompileOutput) or Error (PrecompileError enum).
- `ForSpec(forkID)` — Returns cached `*PrecompileSet` for the fork.

## Gotchas

- gnark-crypto `SetBigInt()` silently reduces mod p — must use `SetBytesCanonical()` for BN254/BLS input validation.
- MODEXP: `adjustedExponentLength` must return `max(result, 1)` (upstream's `max(iteration_count, 1)`).

## Dependencies

`internal/primitives`, `internal/spec`, gnark-crypto, go-kzg-4844
