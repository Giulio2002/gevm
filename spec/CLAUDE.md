# spec

Fork definitions, gas constants, and pre-computed gas parameter tables.

## Key types

- **ForkID** (uint8) — 21 Ethereum hard forks: `Frontier`(0) through `Amsterdam`(20). `LatestForkID = Osaka`.
- **GasParams** — Pre-computed [39]uint64 gas table indexed by `GasId`. One cached instance per ForkID; `NewGasParams(forkID)` returns a shared pointer.
- **GasId** — Enum (1–39) for parameterized gas costs (SSTORE variants, cold/warm access, call stipend, etc.).

## Key functions

- `ForkID.IsEnabledIn(other)` — True if `other` fork is active (i.e., `self >= other`).
- `ForkIDFromString(s)` — Parse fork name from spec test JSON.
- `NewGasParams(forkID)` — Returns cached `*GasParams` for the fork.
- `GasParams.SstoreDynamicGas(...)`, `SelfdestructCost(...)`, etc. — Fork-aware gas calculators.

## Constants

Gas constants: `GasExp=10`, `GasKeccak256=30`, `GasWarmStorageReadCost=100`, `GasColdSloadCost=2100`, `GasIstanbulSloadGas=800`, etc.

## Dependencies

`internal/primitives`
