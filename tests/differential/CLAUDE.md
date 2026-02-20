# tests/differential

Differential testing harness: golden fixtures, determinism, random bytecode, cross-fork consistency.

## Test types

- **Golden fixtures** — Crafted bytecodes with exact expected gas/output/state.
- **Determinism** — Run same input twice, verify identical results.
- **Random bytecode** — 200 random opcode sequences, verify deterministic execution.
- **Cross-fork consistency** — ADD returns same result Berlin→Prague.
- **Revme comparison** — When `REVME_BIN` env var set, compares JSON output against upstream.

## Running

```
go test ./tests/differential/... -count=1
REVME_BIN=/path/to/revme go test ./tests/differential/... -count=1  # with upstream comparison
```

## Dependencies

`internal/host`, `internal/primitives`, `internal/spec`, `tests/spec`
