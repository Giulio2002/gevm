# Glamsterdam / Amsterdam: experimental execution support

**This is not full Glamsterdam support.** The official execution fork name is
`Amsterdam`; `LatestForkID` deliberately remains `Osaka`. Enabling Amsterdam
currently opts into incomplete execution semantics and mostly inherited Osaka
gas accounting. Do not use it for consensus or production transaction execution.

## Implemented and regression-tested here

- Allocate both cached gas tables through Amsterdam (previously selecting the
  existing Amsterdam enum panicked).
- EIP-8024: update DUPN/SWAPN immediate decoding to `(x + 145) % 256`;
  EXCHANGE to the XOR-143 mapping, including valid immediates 80 and 81;
  report stack underflow separately from overflow. Test all 256 immediate bytes
  for each opcode in both generated execution modes, plus gas, fork-gating,
  stack-limit, truncated-immediate, and legacy JUMPDEST behavior.
- EIP-7843 fixture integration: propagate `env.slotNumber` into `BlockEnv.SlotNum`.
  SLOTNUM itself already existed behind the Amsterdam fork gate.
- EIP-7954 limits: 65,536-byte runtime code and 131,072-byte initcode; consistent
  enforcement in top-level creation, CREATE/CREATE2 and transaction-fixture
  validation. Osaka limits are unchanged. This does **not** implement the new
  state-gas cost of deploying that code.
- EIP-7708 execution transfer logs: CALL/top-level value transfers, creation
  endowments and SELFDESTRUCT transfers use the system emitter and Transfer
  topic. Zero/self transfers do not emit; frame rollback discards logs. Logs
  are not charged as LOG3 instructions. Gas payments/refunds are not transfers.
  Withdrawal logs and other block processing are not implemented here.
- EIP-8246 journal semantics: SELFDESTRUCT to self no longer burns a newly
  created contract's balance. `Journal.CommitTx` clears destroyed code, nonce
  and storage while preserving any remaining balance, including late funding.
  Checkpoint rollback and pre-Amsterdam behavior are regression-tested.

These component tests **do not imply that corresponding complete official
state-test groups pass**. Those tests also depend on transaction/gas semantics.
For example all 288 EIP-8024 official state cases still fail full state-root
validation despite the corrected, exhaustively tested instruction behavior.

## Remaining execution work

The release is substantially newer than the old Amsterdam scaffolding:

- EIP-8037 execution/state gas dimensions, state reservoir, spill repayment,
  frame rollback/commit accounting, account/storage/code/authorization charges,
  and execution versus total transaction limits are not implemented. This is a
  structural change, not safely approximated by increasing scalar gas constants.
- EIP-2780 intrinsic/value/recipient/authorization accounting is not implemented.
- EIP-7976 calldata-floor repricing and EIP-7981 access-list repricing are not
  implemented; EIP-8038 state-access/write repricing is also missing.
- EIP-7997 deterministic factory installation is not implemented.
- Block-level requirements (EIP-7778 block gas accounting, EIP-7928 block access
  lists, EIP-8070 sparse blobpool, EIP-8282 builder requests, fork transitions and
  BPO blob schedules) are not validated by this EVM-library state-test run.
  Passing a state case located under a block-access-list directory is **not**
  evidence of block-access-list support.
- Exact transaction exception matching is still missing in the legacy fixture
  executor. Strict reports mark every `expectException` case unsupported rather
  than accepting an unrelated validation error. Blockchain/engine/transaction
  fixture suites were not run as part of the reported totals.
- The separate legacy `RunTestFileOutcomes` and blockchain harnesses are not the
  strict reporting path introduced here. Use `RunFixtureReport` for these totals.

The legacy post-state account-dump comparison also remains a diagnostic aid,
not a substitute for a full root. In particular, callers must finalize journal
selfdestructs when persisting state; this patch does not turn GEVM into a complete
block processor.

## Fixture provenance

Verified via GitHub `releases/latest` on October 28, 2026:

- Repository: `ethereum/execution-specs`
- Release: `tests@v21.0.2`, published `2026-10-26T16:29:51Z`
- Source commit: `e27f097ec7b3e6e8ee9c2c1873c183e8f7da7a6e`
- Asset: `fixtures.tar.gz`, 939,189,561 bytes
- SHA-256, checked against the downloaded bytes:
  `e2fa7ea2ea516459f9ab92284b2927892f318f4e1f78ca802cb05a6c3b479ed9`
- Reference semantics: `src/ethereum/forks/amsterdam/` in that source commit.

The archive is not committed. One unchanged Osaka fixture is retained under
`tests/spec/testdata/eest-v21.0.2/` to prove strict MPT validation without a network
or external fixture installation. Its `_info` retains upstream provenance.

## Strict fixture validation and observed results

`RunFixtureReport` enumerates every state post/fork/index case, counting passed,
failed, unsupported and skipped independently. Missing/empty/malformed inputs
are errors. Unknown forks and cases missing expected roots are unsupported.
No filename skip list is applied by `TestEESTFixtures`.

Successful cases validate the complete merged pre-state/journal MPT root and
logs hash, not just listed accounts. Enabling this uncovered an old MPT helper
bug: leaf payloads must be RLP byte strings, rather than embedded account lists.
The bundled official root is an independent regression oracle. Incorrect hashes
are rejected even if the account dump matches.

| Scope | Files | Cases | Passed | Failed | Unsupported | Skipped |
|---|---:|---:|---:|---:|---:|---:|
| `state_tests/for_osaka` | 2,190 | 14,665 | 13,850 | 0 | 815 | 0 |
| `state_tests/for_amsterdam` | 2,717 | 16,908 | 1,131 | 14,742 | 1,035 | 0 |

Each total was reconciled against an independent enumeration of the downloaded
JSON `post` arrays. No execution panics occurred in the Amsterdam run.
Per-feature breakdown: [glamsterdam-fixture-results.json](glamsterdam-fixture-results.json).
Both entire-suite test commands correctly exit nonzero: unsupported is not pass.
The Amsterdam failure set includes 9,052 unexpected validation errors, 5,587
state-root mismatches and 103 logs-hash mismatches.

## Reproduce

From the repository root, choose an external working directory with enough disk
space. These commands pin the release rather than fetching an unversioned URL:

```sh
FIXTURE_WORK="$HOME/.cache/gevm-eest-v21.0.2"
mkdir -p "$FIXTURE_WORK"
curl --fail --location --retry 3 \
  'https://github.com/ethereum/execution-specs/releases/download/tests%40v21.0.2/fixtures.tar.gz' \
  -o "$FIXTURE_WORK/fixtures.tar.gz"
shasum -a 256 "$FIXTURE_WORK/fixtures.tar.gz"
# Stop unless this matches the SHA-256 above.
tar -xzf "$FIXTURE_WORK/fixtures.tar.gz" -C "$FIXTURE_WORK" fixtures/state_tests

GEVM_EEST_DIR="$FIXTURE_WORK/fixtures/state_tests/for_osaka" \
GEVM_EEST_REPORT="$FIXTURE_WORK/osaka-report.json" \
go test ./tests/spec -run '^TestEESTFixtures$' -v -count=1 -timeout=30m

GEVM_EEST_DIR="$FIXTURE_WORK/fixtures/state_tests/for_amsterdam" \
GEVM_EEST_REPORT="$FIXTURE_WORK/amsterdam-report.json" \
go test ./tests/spec -run '^TestEESTFixtures$' -v -count=1 -timeout=30m
```

`GEVM_EEST_DIR` should be absolute (Go runs package tests in the package directory).
`GEVM_EEST_REPORT` records every case and its status before the test fails.

Local regression checks actually run on Go 1.25.5, darwin/arm64:

```sh
go test ./... -count=1
go test -race ./...
go vet ./...
go generate ./vm
git diff --check
```

Unit/regression suites and vet passed. Race tests passed with macOS linker
`LC_DYSYMTAB` warnings; no race failures were reported. Regenerating the opcode
dispatch produced no generated-code diff. External fixture tests are opt-in and
are skipped by an ordinary `go test ./...` without `GEVM_EEST_DIR`.
