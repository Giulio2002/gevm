package host

import (
	"github.com/Giulio2002/gevm/spec"
	"github.com/Giulio2002/gevm/types"
	"github.com/holiman/uint256"
)

var transferLogAddress = types.Address{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xfe}
var transferLogTopic = types.Keccak256([]byte("Transfer(address,address,uint256)"))

// emitTransferLog records an EIP-7708 execution value transfer. The caller
// must have a journal checkpoint, so failed/reverted frames discard the log.
// Gas purchases, refunds, beneficiary rewards and self-transfers are excluded.
func (h *EvmHost) emitTransferLog(from, to types.Address, amount uint256.Int) {
	if !h.Journal.Cfg.Spec.IsEnabledIn(spec.Amsterdam) || amount.IsZero() || from == to {
		return
	}
	log := h.Journal.AllocLog()
	log.Address = transferLogAddress
	log.NumTopics = 3
	log.Topics = [4]types.B256{transferLogTopic}
	copy(log.Topics[1][12:], from[:])
	copy(log.Topics[2][12:], to[:])
	data := amount.Bytes32()
	log.Data = data[:]
}
