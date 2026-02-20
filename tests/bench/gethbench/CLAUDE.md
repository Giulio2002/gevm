# tests/bench/gethbench

Separate Go module providing go-ethereum benchmarks with identical names to GEVM's `tests/bench/` for automated benchstat comparison.

## Running

```bash
cd tests/bench/gethbench && go test -bench=. -benchmem -count=5
```

## Architecture

- **Separate `go.mod`**: Depends on `github.com/ethereum/go-ethereum v1.16.8`. Isolated from the main GEVM module to avoid pulling geth into the dependency tree.
- **`runtime.Call`**: Uses go-ethereum's `core/vm/runtime` package for execution (not full `Transact`).
- **Shared testdata**: Symlinked `testdata/` with `analysis.hex`, `snailtracer.hex`, `erc20_runtime.hex`.

## Benchmarks

| Benchmark | Description |
|-----------|-------------|
| SWAP1/10k | 10,000 SWAP1 ops |
| RETURN/{1K,10K,100K,1M} | Memory return of various sizes |
| SimpleLoop/* | Tight loops: staticcall/call identity, loop, call-nonexist (100M gas) |
| CREATE_500/1200, CREATE2_500/1200 | Contract creation loops |
| Transfer | Simple ETH value transfer |
| Analysis | ERC-20 deployment code execution |
| Snailtracer | Compute-heavy ray tracer |
| ERC20Transfer | Token transfer with storage reads/writes |
| TenThousandHashes | 10,000 sequential keccak256 |
| Opcode/* | Individual opcode tight loops (ADD, MUL, DIV, KECCAK256, etc.) |

## Dependencies

`github.com/ethereum/go-ethereum` (external)
