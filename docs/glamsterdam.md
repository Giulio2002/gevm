# Glamsterdam / Amsterdam state execution

`spec.Amsterdam` passes **all 16,908 official Amsterdam state-test cases** in
`ethereum/execution-specs tests@v21.0.2`, with complete post-state MPT roots,
logs hashes, and exact transaction-rejection identities checked. The same run
passes all 14,665 Osaka cases. Both the default and opcode-tracing runners pass.

The default `LatestForkID` remains **Osaka**. These results establish state-test
conformance to the pinned release, **not complete Glamsterdam block/engine or
network conformance**. Amsterdam remains an explicit opt-in.

## Execution changes

- **EIP-8037:** distinct execution gas and state reservoirs, execution grants
  capped independently of total transaction gas, reservoir-first charging,
  execution spill tracking, LIFO state refunds, complete reservoir forwarding
  to child frames, cross-frame spill repayment, and rollback before exceptional
  execution-gas forfeiture. Reverted state charges refill; committed
  authorization charges survive dispatched-code failure. Preparation failure
  instead rolls back all authorizations and restores state gas to frame entry.
- **State creation:** SSTORE zero-slot creation/restoration, CALL recipient
  creation, top-level and nested CREATE/CREATE2, code deposition, authorization
  accounts/delegations, and SELFDESTRUCT beneficiary creation. CREATE charges
  state gas before calculating its EIP-150 child execution grant. Deployment
  charges code-hash execution gas separately from per-byte state gas.
- **EIP-2780:** decomposed sender, recipient, value, creation, and authorization
  intrinsic costs; first-write accounting and sender/value-recipient exemptions
  for authorization writes. Authorization nonce changes are journaled so an
  out-of-gas preparation can undo earlier tuples.
- **EIP-7976 / EIP-7981 / EIP-8038:** uniform calldata-floor pricing, access-list
  data surcharges and entry costs, cold account pricing, storage write/refund
  pricing, CALL value pricing, and EXTCODESIZE/EXTCODECOPY code-read surcharges.
- **Gas settlement:** `ExecutionResult.GasUsed` remains the sender-facing amount
  after refund capping and the floor. Amsterdam additionally exposes
  `ExecutionGasUsed` and `StateGasUsed`, using the pre-refund block-accounting
  dimensions. These fields are zero for older forks.
- **EIP-8024:** corrected DUPN/SWAPN and EXCHANGE immediate mappings, with all
  256 immediates checked in both execution modes. Stack errors, fork gates,
  truncated immediates, gas and JUMPDEST behavior have local regressions.
- **EIP-7843 / EIP-7954:** fixture slot-number propagation and Amsterdam's
  65,536-byte runtime / 131,072-byte initcode limits.
- **EIP-7708 / EIP-8246:** execution transfer logs and SELFDESTRUCT finalization
  preserving balance while clearing destroyed code, nonce and storage. The
  optional fixture account dump is checked after transaction finalization,
  matching the already-validated full state root.
- **Validation:** distinguish intrinsic-floor and block-gas errors; enforce
  maximum nonce and Osaka-onward six-blob transaction limit. The strict fixture
  path checks signed-envelope chain IDs, signature bounds and recovery for
  rejection cases, matches the exact allowed exception identity, and checks
  rejection state roots and empty logs. Unknown exception identities remain
  unsupported, never silently accepted.

The state reservoir is a frame-local accounting change, not a scalar gas-cost
approximation. A small bundled regression set exercises spill repayment,
ancestor rollback, authorization preparation failure, committed authorization
charges, selfdestruct finalization and rejected transactions without downloading
external fixtures.

## Scope boundaries

- This library does not become a full block processor by passing state tests.
  Block gas-pool enforcement, block access-list construction/validation
  (EIP-7928), withdrawal logs, builder requests (EIP-8282), sparse blobpool
  behavior (EIP-8070), fork transitions and BPO schedules are not established by
  this run. Per-transaction gas dimensions are exposed, but applying them across
  a complete block is the integrator's responsibility.
- EIP-7997 state cases run successfully against fixture-provided pre-state.
  Automatic deterministic-factory installation at a chain fork transition is
  not implemented here; passing those cases does not establish installation.
- The host consumes already-recovered transaction objects. Signed-envelope
  validation for rejection fixtures lives in the fixture harness, not the host.
- Use **`RunFixtureReport`** for the reported conformance results. The legacy
  `RunTestFileOutcomes`, blockchain and standalone transaction-test paths are
  separate and have not been upgraded into an equivalent strict reporting path.
- Callers must finalize the journal with `Journal.CommitTx` when persisting a
  completed transaction, including Amsterdam SELFDESTRUCT cleanup.
- No blockchain, engine or standalone transaction fixture-suite totals are
  claimed for this release. No remaining failures or exclusions were observed
  in either of the two state-test scopes below.

## Fixture provenance

Pinned release, verified on October 28, 2026:

- Repository: `ethereum/execution-specs`
- Release: `tests@v21.0.2`, published `2026-10-26T16:29:51Z`
- Source commit: `e27f097ec7b3e6e8ee9c2c1873c183e8f7da7a6e`
- Reference semantics: `src/ethereum/forks/amsterdam/` at that exact commit
- Asset: `fixtures.tar.gz`, 939,189,561 bytes
- Downloaded asset SHA-256:
  `e2fa7ea2ea516459f9ab92284b2927892f318f4e1f78ca802cb05a6c3b479ed9`

The full archive is external. Ten unchanged Amsterdam fixture files (18 cases)
and one unchanged Osaka fixture are retained under
`tests/spec/testdata/eest-v21.0.2/`. The upstream `_info` fields remain intact.
The results JSON records bundled file hashes and source paths.

## Verified results

| Scope | Files | Cases | Passed | Failed | Unsupported | Skipped |
|---|---:|---:|---:|---:|---:|---:|
| `state_tests/for_osaka` | 2,190 | 14,665 | 14,665 | 0 | 0 | 0 |
| `state_tests/for_amsterdam` | 2,717 | 16,908 | 16,908 | 0 | 0 | 0 |

Both rows pass independently with `GEVM_TEST_RUNNER=default` and `opcode`.
Counts are reconciled against a separate enumeration of every downloaded JSON
`post[fork][index]` case; no filename exclusions are applied. The Amsterdam
scope includes 1,035 rejection cases and Osaka includes 815, now checked rather
than excluded. Per-feature counts and provenance are recorded in
[glamsterdam-fixture-results.json](glamsterdam-fixture-results.json).

## Reproduce

```sh
FIXTURE_WORK="$HOME/.cache/gevm-eest-v21.0.2"
mkdir -p "$FIXTURE_WORK"
curl --fail --location --retry 3 \
  'https://github.com/ethereum/execution-specs/releases/download/tests%40v21.0.2/fixtures.tar.gz' \
  -o "$FIXTURE_WORK/fixtures.tar.gz"
shasum -a 256 "$FIXTURE_WORK/fixtures.tar.gz"
# Stop unless this matches the SHA-256 above.
tar -xzf "$FIXTURE_WORK/fixtures.tar.gz" -C "$FIXTURE_WORK" fixtures/state_tests

for runner in default opcode; do
  for fork in osaka amsterdam; do
    GEVM_TEST_RUNNER="$runner" \
    GEVM_EEST_DIR="$FIXTURE_WORK/fixtures/state_tests/for_$fork" \
    GEVM_EEST_REPORT="$FIXTURE_WORK/$fork-$runner-report.json" \
    go test ./tests/spec -run '^TestEESTFixtures$' -v -count=1 -timeout=30m || exit 1
  done
done
```

Use absolute fixture paths: Go runs package tests from the package directory.
Reports are written before the suite returns, including on failure. Missing,
empty or malformed fixture inputs fail instead of producing a zero-case pass.

Local checks on Go 1.25.5, darwin/arm64:

```sh
go test ./... -count=1
go test -race ./...
go vet ./...
go generate ./vm
git diff --check
```

All passed. Race builds emit macOS `LC_DYSYMTAB` linker warnings, with no race
failures. Code generation produces no generated-code changes. The complete
external suites require `GEVM_EEST_DIR`; the bundled regressions run in ordinary
`go test ./...` and exercise both execution modes.
