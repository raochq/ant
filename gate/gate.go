package gate

import (
	"github.com/raochq/ant/config"
	"github.com/raochq/ant/protocol/pb"
	"github.com/raochq/ant/service"
	"github.com/raochq/ant/util/logger"
)

type Gate struct {
	id     uint32
	zoneID uint32
	config *pb.GateConfig
	name   string

	state service.State
}

var _ service.IService = (*Gate)(nil)

func (g *Gate) StateInfo() interface{} {
	return nil
}

func (g *Gate) Stop() {
}

func (g *Gate) Init() error {
	logger.Info("Gate init...\n")
	return nil
}
func (g *Gate) Close() {
	logger.Info("Gate destroy...\n")
}

func (g *Gate) MainLoop(sig <-chan byte) {
	logger.Info("Gate Run in Loop\n")
}

func New(name string, info pb.ServiceInfo) *Gate {
	if info.GateConfig == nil {
		return nil
	}
	return &Gate{
		name:   name,
		id:     info.ID,
		zoneID: info.Zone,
		config: info.GateConfig,
	}
}
func init() {
	service.Register(pb.ServiceKind_Gate.String(), func(conf config.Config) service.IService {
		//return New(name, info)
		return nil
	})
}
