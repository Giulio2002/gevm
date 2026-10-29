// Provides a concrete Host implementation that bridges the Journal/state
// system to the VM's Host interface.
package host

import (
	"github.com/Giulio2002/gevm/precompiles"
	"github.com/Giulio2002/gevm/state"
	"github.com/Giulio2002/gevm/types"
	"github.com/Giulio2002/gevm/vm"
	"github.com/holiman/uint256"
)

// BlockEnv holds block-level environment data.
type BlockEnv struct {
	Beneficiary  types.Address
	Timestamp    uint256.Int
	Number       uint256.Int
	Difficulty   uint256.Int
	Prevrandao   *uint256.Int
	GasLimit     uint256.Int
	BaseFee      uint256.Int
	BlobGasPrice uint256.Int
	SlotNum      uint256.Int
	GetHash      func(uint64) (types.B256, error)
}

// TxEnv holds transaction-level environment data.
type TxEnv struct {
	Caller            types.Address
	EffectiveGasPrice uint256.Int
	BlobHashes        []uint256.Int
}

// CfgEnv holds chain configuration.
type CfgEnv struct {
	ChainId uint256.Int
}

// EvmHost is the concrete Host implementation that delegates to Journal.
// It implements the vm.Host interface.
// Block and Cfg are stored by pointer to avoid ~316 bytes of struct copy
// when setting up the host in Transact.
type EvmHost struct {
	Block   *BlockEnv
	Tx      TxEnv
	Cfg     *CfgEnv
	Journal *state.Journal

	Precompiles               *precompiles.PrecompileSet
	DisablePrecompileFastPath bool
	Hooks                     *vm.Hooks
}

// NewEvmHost creates a new EvmHost.
func NewEvmHost(journal *state.Journal, block *BlockEnv, tx TxEnv, cfg *CfgEnv) *EvmHost {
	return &EvmHost{
		Block:   block,
		Tx:      tx,
		Cfg:     cfg,
		Journal: journal,
	}
}

// Compile-time check that EvmHost implements vm.Host.
var _ vm.Host = (*EvmHost)(nil)

// --- Block info ---

func (h *EvmHost) Beneficiary() types.Address { return h.Block.Beneficiary }
func (h *EvmHost) Timestamp() uint256.Int     { return h.Block.Timestamp }
func (h *EvmHost) BlockNumber() uint256.Int   { return h.Block.Number }
func (h *EvmHost) Difficulty() uint256.Int    { return h.Block.Difficulty }
func (h *EvmHost) Prevrandao() *uint256.Int   { return h.Block.Prevrandao }
func (h *EvmHost) GasLimit() uint256.Int      { return h.Block.GasLimit }
func (h *EvmHost) BaseFee() uint256.Int       { return h.Block.BaseFee }
func (h *EvmHost) BlobGasPrice() uint256.Int  { return h.Block.BlobGasPrice }
func (h *EvmHost) SlotNum() uint256.Int       { return h.Block.SlotNum }
func (h *EvmHost) ChainId() uint256.Int       { return h.Cfg.ChainId }

// --- Tx info ---

func (h *EvmHost) Caller() types.Address          { return h.Tx.Caller }
func (h *EvmHost) EffectiveGasPrice() uint256.Int { return h.Tx.EffectiveGasPrice }
func (h *EvmHost) BlobHash(index int) *uint256.Int {
	if index < 0 || index >= len(h.Tx.BlobHashes) {
		return nil
	}
	v := h.Tx.BlobHashes[index]
	return &v
}

// --- Account access (delegated to Journal) ---

func (h *EvmHost) Balance(addr types.Address) (uint256.Int, bool) {
	result, err := h.Journal.LoadAccount(addr)
	if err != nil {
		return uint256.Int{}, false
	}
	return result.Data.Info.Balance, result.IsCold
}

func (h *EvmHost) CodeSize(addr types.Address) (int, bool) {
	result, err := h.Journal.LoadAccount(addr)
	if err != nil {
		return 0, false
	}
	acc := result.Data
	code, dbErr := h.loadCode(addr, acc)
	if dbErr != nil {
		return 0, false
	}
	return len(code), result.IsCold
}

func (h *EvmHost) CodeHash(addr types.Address) (types.B256, bool) {
	result, err := h.Journal.LoadAccount(addr)
	if err != nil {
		return types.B256Zero, false
	}
	acc := result.Data
	if acc.StateClearAwareIsEmpty(h.Journal.Cfg.Spec) {
		return types.B256Zero, result.IsCold
	}
	return acc.Info.CodeHash, result.IsCold
}

func (h *EvmHost) Code(addr types.Address) (types.Bytes, bool) {
	result, err := h.Journal.LoadAccount(addr)
	if err != nil {
		return nil, false
	}
	acc := result.Data
	code, dbErr := h.loadCode(addr, acc)
	if dbErr != nil {
		return nil, false
	}
	return code, result.IsCold
}

func (h *EvmHost) LoadAccountCode(addr types.Address) vm.AccountCodeLoad {
	result, err := h.Journal.LoadAccount(addr)
	if err != nil {
		return vm.AccountCodeLoad{IsEmpty: true}
	}
	acc := result.Data
	isEmpty := acc.StateClearAwareIsEmpty(h.Journal.Cfg.Spec)
	code, _ := h.loadCode(addr, acc)
	return vm.AccountCodeLoad{
		Code:     code,
		CodeHash: acc.Info.CodeHash,
		IsCold:   result.IsCold,
		IsEmpty:  isEmpty,
	}
}

func (h *EvmHost) loadCode(addr types.Address, acc *state.Account) (types.Bytes, error) {
	if acc.Info.CodeHash == types.KeccakEmpty || acc.Info.CodeHash.IsZero() {
		acc.Info.Code = nil
		return nil, nil
	}
	if acc.Info.Code != nil {
		return acc.Info.Code, nil
	}
	code, err := h.Journal.ReadCode(addr)
	if err != nil {
		return nil, err
	}
	acc.Info.Code = code
	return code, nil
}

// IsPrecompile returns true if addr is a precompile for the active fork.
func (h *EvmHost) IsPrecompile(addr types.Address) bool {
	return !h.DisablePrecompileFastPath && h.Precompiles != nil && h.Precompiles.Get(addr) != nil
}

// RunPrecompile executes addr if it is a precompile for the active fork.
func (h *EvmHost) RunPrecompile(addr types.Address, input types.Bytes, gasLimit uint64) (vm.PrecompileCallResult, bool) {
	if h.Precompiles == nil {
		return vm.PrecompileCallResult{}, false
	}
	precompile := h.Precompiles.Get(addr)
	if precompile == nil {
		return vm.PrecompileCallResult{}, false
	}
	if len(input) == 0 && isIdentityPrecompileAddress(addr) {
		if gasLimit < 15 {
			return vm.PrecompileCallResult{Result: vm.InstructionResultPrecompileOOG}, true
		}
		return vm.PrecompileCallResult{
			Result:  vm.InstructionResultReturn,
			GasUsed: 15,
		}, true
	}

	execResult := precompile.Execute(input, gasLimit)
	if execResult.IsErr() {
		resultCode := vm.InstructionResultPrecompileError
		if *execResult.Err == precompiles.PrecompileErrorOutOfGas {
			resultCode = vm.InstructionResultPrecompileOOG
		}
		return vm.PrecompileCallResult{Result: resultCode}, true
	}

	output := execResult.Output
	resultCode := vm.InstructionResultReturn
	if output.Reverted {
		resultCode = vm.InstructionResultRevert
	}
	return vm.PrecompileCallResult{
		Result:    resultCode,
		Output:    output.Bytes,
		GasUsed:   output.GasUsed,
		GasRefund: output.GasRefund,
	}, true
}

func (h *EvmHost) PrecompileHooks() *vm.Hooks {
	return h.Hooks
}

func isIdentityPrecompileAddress(addr types.Address) bool {
	for i := 0; i < len(addr)-1; i++ {
		if addr[i] != 0 {
			return false
		}
	}
	return addr[19] == 0x04
}

func (h *EvmHost) SelfBalance(addr types.Address) uint256.Int {
	if acc, ok := h.Journal.State[addr]; ok {
		return acc.Info.Balance
	}
	return uint256.Int{}
}

// --- Storage access ---

func (h *EvmHost) SLoadInto(addr types.Address, key *uint256.Int, out *uint256.Int) bool {
	isCold, err := h.Journal.SLoadInto(addr, key, out)
	if err != nil {
		*out = uint256.Int{}
		return false
	}
	return isCold
}

func (h *EvmHost) SStore(addr types.Address, key *uint256.Int, value *uint256.Int, out *vm.SStoreResult) {
	err := h.Journal.SStoreInto(addr, key, value,
		&out.OriginalValue, &out.PresentValue, &out.NewValue, &out.IsCold)
	if err != nil {
		*out = vm.SStoreResult{}
	}
}

func (h *EvmHost) TLoad(addr types.Address, key uint256.Int) uint256.Int {
	return h.Journal.TLoad(addr, key)
}

func (h *EvmHost) TStore(addr types.Address, key, value uint256.Int) {
	h.Journal.TStore(addr, key, value)
}

// --- Block hash ---

func (h *EvmHost) BlockHash(number uint256.Int) types.B256 {
	if number[1] != 0 || number[2] != 0 || number[3] != 0 {
		return types.B256Zero
	}
	n := number[0]
	current := h.Block.Number[0]
	if h.Block.Number[1] != 0 || h.Block.Number[2] != 0 || h.Block.Number[3] != 0 || n >= current || current-n > 256 {
		return types.B256Zero
	}
	if h.Block != nil && h.Block.GetHash != nil {
		hash, err := h.Block.GetHash(n)
		if err != nil {
			return types.B256Zero
		}
		return hash
	}
	if h.Journal.DB == nil {
		return types.B256Zero
	}
	hash, err := h.Journal.BlockHash(n)
	if err != nil {
		return types.B256Zero
	}
	return hash
}

// --- Logging ---

func (h *EvmHost) Log(addr types.Address, topics *[4]types.B256, numTopics int, data types.Bytes) {
	log := h.Journal.AllocLog()
	log.Address = addr
	log.Topics = *topics
	log.NumTopics = uint8(numTopics)
	log.Data = data
}

// --- Self destruct ---

func (h *EvmHost) SelfDestruct(addr, target types.Address) vm.SelfDestructResult {
	balance := h.Journal.State[addr].Info.Balance
	result, err := h.Journal.Selfdestruct(addr, target)
	if err != nil {
		return vm.SelfDestructResult{}
	}
	h.emitTransferLog(addr, target, balance)
	return vm.SelfDestructResult{
		HadValue:            result.Data.HadValue,
		TargetExists:        result.Data.TargetExists,
		IsCold:              result.IsCold,
		PreviouslyDestroyed: result.Data.PreviouslyDestroyed,
	}
}
