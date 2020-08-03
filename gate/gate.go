package gate

import (
	"github.com/raochq/ant/engine/logger"
	"github.com/raochq/ant/protocol/pb"
	"github.com/raochq/ant/service"
)

type Gate struct {
	pb.ServiceInfo
	name string

	state service.State
}

func (g *Gate) Stop() {
}

func (g *Gate) Init() error {
	logger.Info("Gate init...\n")
	return nil
}
func (g *Gate) Destroy() {
	logger.Info("Gate destroy...\n")
}
func (g *Gate) MainLoop(sig <-chan byte) {
	logger.Info("Gate Run in Loop\n")
}

func New(name string, info pb.ServiceInfo) *Gate {
	return &Gate{
		name:        name,
		ServiceInfo: info,
	}
}
func init() {
	service.Register(pb.ServiceInfo_Gate, func(name string, info pb.ServiceInfo) service.IService {
		return New(name, info)
	})
}
