# opcode

EVM opcode constants and metadata. 154 valid opcodes as `byte` constants.

## Key exports

- **Opcode constants** — `STOP`(0x00) through `SELFDESTRUCT`(0xFF). PUSH0–PUSH32, DUP1–DUP16, SWAP1–SWAP16, LOG0–LOG4.
- **OpcodeInfo** — [256]*Info table with name, stack inputs/outputs, terminating flag.
- `IsValid(op)`, `IsPush(op)`, `IsTerminating(op)`, `ImmediateSize(op)`, `Name(op)`.

## Notes

- No gas tables here — fork-varying gas is in `vm.ForkGas`; static gas is hardcoded in `vm/table.go` switch cases.
- This package has zero dependencies.
