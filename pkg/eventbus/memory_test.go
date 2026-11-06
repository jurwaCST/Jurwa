package eventbus_test

import (
	"testing"

	"github.com/jurwaCST/Jurwa/pkg/eventbus"
)

func TestMemoryBus(t *testing.T) {
	runBusContract(t, func(*testing.T) Bus { return eventbus.NewMemory() })
}
