package huron

import (
	_huron "github.com/abassian/huron/src/huron"
	"github.com/abassian/shuffle/src/config"
	"github.com/abassian/shuffle/src/service"
	"github.com/abassian/shuffle/src/state"
	"github.com/sirupsen/logrus"
)

// InmemHuron implementes the Consensus interface.
// It uses an inmemory Huron node.
type InmemHuron struct {
	config     *config.HuronConfig
	huron     *_huron.Huron
	ethService *service.Service
	ethState   *state.State
	logger     *logrus.Logger
}

// NewInmemHuron instantiates a new InmemHuron consensus system
func NewInmemHuron(config *config.HuronConfig, logger *logrus.Logger) *InmemHuron {
	return &InmemHuron{
		config: config,
		logger: logger,
	}
}

/*******************************************************************************
IMPLEMENT CONSENSUS INTERFACE
*******************************************************************************/

// Init instantiates a Huron inmemory node
func (b *InmemHuron) Init(state *state.State, service *service.Service) error {
	b.logger.Debug("INIT")

	b.ethState = state
	b.ethService = service

	realConfig := b.config.ToRealHuronConfig()
	realConfig.Proxy = NewInmemProxy(state, service, service.GetSubmitCh(), b.logger)

	huron := _huron.NewHuron(realConfig)

	err := huron.Init()
	if err != nil {
		return err
	}

	b.huron = huron

	return nil
}

// Run starts the Huron node
func (b *InmemHuron) Run() error {
	b.huron.Run()
	return nil
}

// Info returns Huron stats
func (b *InmemHuron) Info() (map[string]string, error) {
	info := b.huron.Node.GetStats()
	info["type"] = "huron"
	return info, nil
}
