package host

import (
	"github.com/Giulio2002/gevm/spec"
	"github.com/Giulio2002/gevm/types"
	"testing"
)

func TestTransferLogClearsReusedTopics(t *testing.T) {
	h, _ := makeHandler(newMockDB(), spec.Amsterdam)
	cp := h.Host.Journal.Checkpoint()
	l := h.Host.Journal.AllocLog()
	for i := range l.Topics {
		for j := range l.Topics[i] {
			l.Topics[i][j] = 0xff
		}
	}
	h.Host.Journal.CheckpointRevert(cp)
	h.Host.emitTransferLog(addr(1), addr(2), u(3))
	l = &h.Host.Journal.Logs[0]
	var zero types.B256
	for _, topic := range l.Topics[1:3] {
		for _, b := range topic[:12] {
			if b != 0 {
				t.Fatal("reused topic prefix was not cleared")
			}
		}
	}
	if l.Topics[3] != zero {
		t.Fatal("unused topic was not cleared")
	}
}
