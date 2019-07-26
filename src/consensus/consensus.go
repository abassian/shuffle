package consensus

import (
	"github.com/abassian/shuffle/src/service"
	"github.com/abassian/shuffle/src/state"
)

// Consensus is the interface that abstracts the consensus system
type Consensus interface {
	Init(*state.State, *service.Service) error
	Run() error
	Info() (map[string]string, error)
}
