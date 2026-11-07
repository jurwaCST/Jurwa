package memory

import (
	"testing"

	"github.com/jurwaCST/Jurwa/services/identity/adapters/out/repotest"
	"github.com/jurwaCST/Jurwa/services/identity/ports"
)

func TestUsersContract(t *testing.T) {
	repotest.Run(t, func(*testing.T) ports.UserRepository { return NewUsers() })
}
