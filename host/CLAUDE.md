# host

Top-level EVM executor, frame handler, and Host interface implementation.

## Key types

- **Evm** — Top-level executor. Owns Journal, Handler, root Memory, pools. `Transact(tx)` runs the full 4-phase pipeline: validation → pre-exec → exec → post-exec.
- **Handler** — Orchestrates frame execution (CALL/CREATE). Routes precompiles, manages sub-interpreters, handles return values.
- **EvmHost** — Implements `vm.Host` interface, bridging Journal state to interpreter queries.
- **BlockEnv** — Beneficiary, timestamp, number, difficulty, prevrandao, gas limit, base fee, blob gas price.
- **TxEnv** — Caller, effective gas price, blob hashes.
- **CfgEnv** — Chain ID.
- **Transaction** — Full tx: kind (call/create), type (legacy/1559/4844), caller, to, value, input, gas, access list, blob hashes.
- **ExecutionResult** — Kind (success/revert/halt), gas used, gas refund, output, logs, ValidationError flag.

## Transaction pipeline

1. **Validate**: tx type-fork gates, gas price, blob validation, EIP-3607, EIP-3860
2. **Pre-exec**: deduct max fee, bump nonce, warm precompiles + coinbase (Shanghai+)
3. **Execute**: `Handler.ExecuteFrame()` with recursive sub-frame handling
4. **Post-exec**: apply refund (EIP-3529: spent/5 London+, spent/2 pre-London), reward beneficiary, EIP-7623 floor gas (Prague+)

## Frame execution

- `executeCall`: checkpoint → value transfer → precompile check → bytecode load → sub-interpreter → run loop
- `executeCreate`: balance check → nonce bump → address calc → CreateAccountCheckpoint → sub-interpreter
- `returnCreate`: EIP-3541 (0xEF prefix), EIP-170 (24KB limit), code deposit cost

## Dependencies

`internal/primitives`, `internal/spec`, `internal/state`, `internal/vm`, `internal/precompiles`, `internal/keccak`
