// gevm-evmbench is a runner for the evm-bench benchmarking framework.
// It deploys a contract via CREATE, then times repeated CALL executions.
// Output format: one line per run with elapsed milliseconds (float).
package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"github.com/holiman/uint256"
	"os"
	"strings"
	"time"

	"github.com/Giulio2002/gevm/host"
	"github.com/Giulio2002/gevm/spec"
	"github.com/Giulio2002/gevm/state"
	spectest "github.com/Giulio2002/gevm/tests/spec"
	"github.com/Giulio2002/gevm/types"
)

func fatal(msg string, args ...any) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(1)
}

func main() {
	contractCodePath := flag.String("contract-code-path", "", "Path to the hex contract code to deploy and run")
	calldata := flag.String("calldata", "", "Hex of calldata to use when calling the contract")
	numRuns := flag.Int("num-runs", 0, "Number of times to run the benchmark")
	flag.Parse()

	if *contractCodePath == "" || *numRuns == 0 {
		fatal("usage: gevm-evmbench --contract-code-path <path> --calldata <hex> --num-runs <n>")
	}

	// Read and decode contract code.
	contractCodeHex, err := os.ReadFile(*contractCodePath)
	if err != nil {
		fatal("reading contract code: %v", err)
	}
	contractCode, err := hex.DecodeString(strings.TrimSpace(string(contractCodeHex)))
	if err != nil {
		fatal("decoding contract hex: %v", err)
	}

	var calldataBytes []byte
	if *calldata != "" {
		calldataBytes, err = hex.DecodeString(*calldata)
		if err != nil {
			fatal("decoding calldata hex: %v", err)
		}
	}

	// Match geth runner's address scheme.
	callerAddress := types.Address{0x10, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0x01}
	coinbase := types.Address{}

	forkID := spec.Cancun
	hugeBalance := uint256.Int{0, 0, 1, 0} // ~2^128
	gasLimit := uint64(1) << 63            // large, won't overflow

	prevrandao := uint256.Int{}
	block := host.BlockEnv{
		Beneficiary: coinbase,
		GasLimit:    *uint256.NewInt(gasLimit),
		BaseFee:     uint256.Int{},
		Number:      *uint256.NewInt(1),
		Timestamp:   *uint256.NewInt(1000),
		Prevrandao:  &prevrandao,
	}
	cfg := host.CfgEnv{ChainId: *uint256.NewInt(1)}

	// Step 1: Deploy contract via CREATE.
	deployDB := spectest.NewMemDB()
	deployDB.InsertAccount(callerAddress, state.AccountInfo{
		Balance:  hugeBalance,
		CodeHash: types.KeccakEmpty,
	}, nil)

	evm := host.NewEvm(deployDB, forkID, block, cfg)
	result := evm.Transact(&host.Transaction{
		Kind:     host.TxKindCreate,
		TxType:   host.TxTypeLegacy,
		Caller:   callerAddress,
		GasLimit: gasLimit,
		Input:    contractCode,
	})

	if !result.IsSuccess() {
		fatal("CREATE failed: result=%d", result.Reason)
	}
	if result.CreatedAddr == nil {
		fatal("CREATE succeeded but no address returned")
	}
	contractAddr := *result.CreatedAddr

	// Extract post-CREATE state into a new MemDB for subsequent calls.
	// Deep-copy Code slices to avoid pool/arena aliasing after ReleaseEvm.
	callDB := spectest.NewMemDB()
	var callerNonce uint64
	for addr, acc := range evm.Journal.State {
		var storage map[uint256.Int]uint256.Int
		if acc.Storage != nil && len(acc.Storage) > 0 {
			storage = make(map[uint256.Int]uint256.Int, len(acc.Storage))
			for key, slot := range acc.Storage {
				storage[key] = slot.PresentValue
			}
		}
		// Deep-copy code to isolate from arena/pool reuse.
		var code types.Bytes
		if len(acc.Info.Code) > 0 {
			code = make(types.Bytes, len(acc.Info.Code))
			copy(code, acc.Info.Code)
		}
		codeHash := acc.Info.CodeHash
		if len(code) == 0 {
			codeHash = types.KeccakEmpty
		}
		callDB.InsertAccount(addr, state.AccountInfo{
			Balance:  acc.Info.Balance,
			Nonce:    acc.Info.Nonce,
			CodeHash: codeHash,
			Code:     code,
		}, storage)
		if addr == callerAddress {
			callerNonce = acc.Info.Nonce
		}
	}
	evm.ReleaseEvm()

	// Step 2: Time CALL executions.
	for i := 0; i < *numRuns; i++ {
		evm := host.NewEvm(callDB, forkID, block, cfg)

		start := time.Now()
		callResult := evm.Transact(&host.Transaction{
			Kind:     host.TxKindCall,
			TxType:   host.TxTypeLegacy,
			Caller:   callerAddress,
			To:       contractAddr,
			GasLimit: gasLimit,
			Nonce:    callerNonce,
			Input:    calldataBytes,
		})
		elapsed := time.Since(start)

		if !callResult.IsSuccess() {
			fmt.Fprintf(os.Stderr, "CALL failed on run %d: result=%d\n", i, callResult.Reason)
		}
		fmt.Println(float64(elapsed.Microseconds()) / 1e3)

		evm.ReleaseEvm()
	}
}
