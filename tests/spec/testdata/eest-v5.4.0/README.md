# EEST v5.4.0 CI regression fixtures

These five JSON files are byte-for-byte copies from the `fixtures_stable.tar.gz`
asset downloaded by `make eest-fixtures` (release `v5.4.0`). No cases, roots,
exception alternatives, or state dumps have been changed. They contain 32 cases:
6 zero-fee touching cases, 15 pre-activation transaction rejections, and 11
post-activation positive controls. The normal full legacy download remains in CI.

The archive's `.meta/fixtures.ini` identifies the actual source as
`ethereum/execution-specs`, ref `refs/tags/tests-v5.4.0`, commit
`4f68564f47c7e577ad6cbb570858316f5ff0e7bb`, filled with `t8n = 2.18.0rc6`.
This source commit, not the old execution-spec-tests repository's same-named
release tag, is the reference for these fixtures.

Paths below are relative to the archive's `fixtures/state_tests/`:

| Source path | SHA-256 (copied bytes) |
| --- | --- |
| `frontier/touch/test_zero_gas_price_and_touching.json` | `3760c6291d0e880d93a67c6c29862d2aa84ecef01a1f1ed170a7e2b874c69028` |
| `berlin/eip2930_access_list/test_eip2930_tx_validity.json` | `53a4a8462c02c65ee31f4c9f8b15f9fe6c635542b14412882ff2561069449232` |
| `london/eip1559_fee_market_change/test_eip1559_tx_validity.json` | `6e8bd4de9b5bda202235265abf8406a847723fc97e702370821c0ff7565d08bb` |
| `cancun/eip4844_blobs/test_blob_type_tx_pre_fork.json` | `f47a4c7d3ca1e750e486549a6aee9c642815448a8dc6f8231b22ee84362e98fb` |
| `prague/eip7702_set_code_tx/test_set_code_type_tx_pre_fork.json` | `7377fb5075117a0125bb28ae05e9b82195f97e1a3b2aafc857dbf51b64e1f6cd` |

## Diagnosis

The Frontier/Homestead post-state includes the empty beneficiary
`0x2adc25665018aa1fe0e6bc666dac8fc2697ff9ba`; the Berlin post-state is otherwise
identical but omits it. GEVM incorrectly avoided touching an absent beneficiary
when the transaction fee was zero even before EIP-161. This produced the Berlin
root `0xe475c3ef3ac00e3f080a340ccb9cca461f0dcff576c014f8c39b64367c7d3396`
instead of the required Frontier/Homestead root
`0x86d5ed0de01cab2622b7d441cc3662d5289a4f008de0585f1def99ebe7dff479`.
The pinned source's `src/ethereum/forks/frontier/fork.py:666-672` unconditionally
sets the beneficiary balance after adding the transaction fee, including zero.
The fix restores this behavior only before Spurious Dragon.

The 15 unsupported cases are valid pre-fork rejections (5 type-1, 6 type-2,
2 type-3, 2 type-4). The host already rejects them with `InvalidTxType`; the
strict fixture adapter lacked the four type-specific `TYPE_n_TX_PRE_FORK`
identities. The adapter now maps the actual rejection and transaction type;
it does not infer success from the expected exception. One blob fixture permits
either `TYPE_3_TX_PRE_FORK` or `TYPE_3_TX_ZERO_BLOBS`; both remain exact matches.

Both runners exercise all 32 pinned cases. Negative controls corrupt the
exception type, state root, and logs hash to verify that the strict checks remain
active. Host tests cover absent-beneficiary touching on both sides of EIP-161.
