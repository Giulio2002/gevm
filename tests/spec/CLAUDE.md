# tests/spec

Ethereum specification test harness. Parses and executes GeneralStateTests, BlockchainTests, and TransactionTests against GEVM.

## Test suites

| Suite | Count | Env var |
|-------|-------|---------|
| GeneralStateTests | 37,692/37,692 (100%) | GEVM_STATE_TESTS_DIR |
| BlockchainTests/ValidBlocks | 882/882 (100%) | GEVM_BLOCKCHAIN_TESTS_DIR |
| BlockchainTests/InvalidBlocks | 260/260 (100%) | GEVM_BLOCKCHAIN_TESTS_DIR |
| TransactionTests | 2,753/2,753 (100%) | GEVM_TRANSACTION_TESTS_DIR |

## Key types

- **TestSuite** — `map[string]*TestUnit`, parsed from JSON fixtures.
- **MemDB** — In-memory `state.Database` implementation for tests.
- **TestOutcome** — Detailed execution result with JSON tags for structured output.
- **RunnerConfig** — Skip list, fork filter, verbosity.

## Key functions

- `RunTestFile(path, cfg)` — Execute all tests in a fixture file.
- `RunTestFileOutcomes(path, cfg)` — Run with detailed JSON outcomes.
- `RunBlockchainTestFile(path, cfg)` — Execute blockchain tests (multi-block, system calls).
- `RunTransactionTestFile(path, cfg)` — Execute transaction validation tests.

## Validation

Every test validates: gas used, logs root (RLP-encoded hash), output data, and post-state (balance, nonce, code, storage). This makes every GST an implicit differential test.

## Key files

- `runner.go` — GST runner, tx builder, state validation
- `blockchain_runner.go` — Multi-block execution, system calls (EIP-2935, EIP-4788, EIP-7002/7251)
- `transaction_runner.go` — RLP decoding, tx validation, sender recovery
- `rlp.go` / `rlp_decode.go` — RLP encode/decode with strict canonical validation
- `outcome.go` — TestOutcome type and JSON serialization

## Dependencies

All `internal/*` packages
