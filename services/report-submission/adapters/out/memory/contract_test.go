package memory

import (
	"testing"

	"github.com/jurwaCST/Jurwa/services/report-submission/adapters/out/repotest"
	"github.com/jurwaCST/Jurwa/services/report-submission/ports"
)

func TestMemoryRepositoryContract(t *testing.T) {
	repotest.RunRepoContract(t, func(*testing.T) ports.ReportRepository { return NewRepository() })
}
