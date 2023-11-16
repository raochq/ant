package gate

import (
	"log/slog"

	"github.com/raochq/ant/config"
	"github.com/raochq/ant/protocol/pb"
	"github.com/raochq/ant/service"
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
	slog.Info("Gate init...")
	return nil
}
func (g *Gate) Close() {
	slog.Info("Gate destroy...")
}

func (g *Gate) MainLoop(sig <-chan byte) {
	slog.Info("Gate Run in Loop")
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
