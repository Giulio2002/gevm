# state

Journaled state: accounts, storage, warm/cold tracking, checkpoints, reverts.

## Key types

- **Journal** — Central state manager. Holds EvmState (map[Address]*Account), transient storage, logs, journal entries, warm addresses, and an account arena (slab allocator).
- **Account** — Info (balance, nonce, code hash, code), original info (for SSTORE refund calc), storage (EvmStorage), status flags.
- **AccountStatus** (uint8) — Bitflags: Created, SelfDestructed, Touched, Cold, LoadedAsNotExist.
- **JournalEntry** — Tagged struct (13 kinds): AccountLoad, BalanceTransfer, NonceChange, CodeChange, StorageChange, etc. Avoids heap allocs.
- **WarmAddresses** — Tracks precompile addresses, coinbase, EIP-2930 access list. First access = cold (higher gas).
- **TransientStorage** — map[Address]map[Uint256]Uint256, cleared per transaction (EIP-1153).
- **Database** (interface) — External state provider (balance, nonce, code, storage queries).

## Key operations

- `Checkpoint() int` / `CheckpointRevert(id)` / `CheckpointCommit()` — Save/restore state.
- `LoadAccount(addr)` — Load from DB, track cold access.
- `Transfer(from, to, amount)` — Balance transfer with journaling.
- `SStore(addr, slot, value)` — Storage write with refund computation.
- `SLoad(addr, slot)` — Storage read with cold/warm tracking.

## Patterns

- Account arena: slab allocator replaces sync.Pool per-account Get/Put.
- 2-entry storage slot cache on Journal for hot SLoad paths.
- `stateAddrs []Address` for O(n) release instead of map iteration.
- Pointer maps (`EvmState`, `EvmStorage`) for in-place mutation.

## Dependencies

`internal/primitives`, `internal/spec`
