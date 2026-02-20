# tests/bench

End-to-end EVM benchmarks exercising the full Transact() pipeline. Benchmark names match `gethbench/` for benchstat comparison.

## Running

```bash
# GEVM benchmarks
go test ./tests/bench/ -bench=. -benchmem -count=5

# Geth benchmarks (separate module)
cd tests/bench/gethbench && go test -bench=. -benchmem -count=5

# Compare with benchstat
benchstat gevm=gevm.txt geth=geth.txt
```

## Benchmarks

| Benchmark | Description |
|-----------|-------------|
| SWAP1/10k | 10,000 SWAP1 ops |
| RETURN/{1K,10K,100K,1M} | Memory return of various sizes |
| SimpleLoop/* | Tight loops: staticcall/call identity, loop, call-nonexist (100M gas) |
| CREATE_500/1200, CREATE2_500/1200 | Contract creation loops |
| Transfer | Simple ETH value transfer (21k gas) |
| Analysis | ERC-20 deployment code execution |
| Snailtracer | Compute-heavy ray tracer (~32ms) |
| ERC20Transfer | Token transfer with storage reads/writes |
| TenThousandHashes | 10,000 sequential keccak256 |
| Opcode/* | Individual opcode tight loops (ADD, MUL, DIV, KECCAK256, etc.) |

## Testdata

Embedded hex files in `testdata/`: `analysis.hex`, `snailtracer.hex`, `erc20_runtime.hex`.

## Dependencies

`internal/host`, `internal/primitives`, `internal/spec`, `internal/state`, `tests/spec`
