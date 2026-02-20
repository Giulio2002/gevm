# vm

EVM interpreter core: bytecode execution, direct switch dispatch, stack/memory/gas, frame actions.

## Key types

- **Interpreter** — Main execution state: Bytecode, Gas, Stack, Memory, ForkGas, GasParams, action/halt fields, scratch buffers (SStoreScratch, TopicsScratch).
- **Bytecode** — Code bytes + PC + jump table + 33-byte end padding for safe PUSH32 reads. Pool-friendly with `Reset()`/`ResetWithHash()`.
- **Stack** — Fixed `[1024]Uint256` array with `top` pointer. No heap.
- **Memory** — Growable byte buffer. Supports checkpoint/restore for nested call contexts.
- **Gas** — `remaining` + `refunded` (int64). Memory expansion tracked via `MemoryGas`.
- **ForkGas** — 6-field struct (48 bytes) caching fork-varying base costs: Balance, ExtCodeSize, ExtCodeHash, Sload, Call, Selfdestruct.
- **InstructionResult** (uint8) — 45 execution outcomes (Stop, Return, Revert, OOG, etc.).
- **Host** (interface) — External environment queries (block, tx, accounts, storage, logs).
- **CallInputs/CreateInputs** — Frame request types.
- **CallOutcome/CreateOutcome/FrameResult** — Frame result types.

## Execution model

`Interpreter.Run(host)` is a direct `switch op` over all 256 byte values. Hot opcodes (ADD, MUL, LT, GT, EQ, AND, OR, MLOAD, MSTORE, JUMP, JUMPI, PUSH0–PUSH4) are inlined in the switch body. Others call `opXxx(interp, host)` helpers in `inst_*.go`.

Fork-varying opcodes read from `interp.ForkGas.X`. All other static gas is hardcoded as immediate constants.

## Object pooling

`sync.Pool` for Stack, Interpreter, Memory, Bytecode, ReturnDataArena. See `pool.go`.

## Stack ordering convention

`push2Op(a, b, op)` → b is on top, op operates on (b, a). CALL stack: retLen(deepest) ... gas(top). CREATE2: salt(deepest) ... value(top).

## Dependencies

`internal/primitives`, `internal/spec`, `internal/opcode`, `internal/keccak`
