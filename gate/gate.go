package gate

import (
	"fmt"
	"github.com/raochq/ant/engine/logger"
	"github.com/raochq/ant/protocol/pb"
	"github.com/raochq/ant/service"
)

type Gate struct {
	zoneId int32
	id     int32
	state  service.State
}

func (g *Gate) Name() string {
	return fmt.Sprintf("/Gate/%d/%d", g.zoneId, g.id)
}
func (g *Gate) ZoneID() int32 {
	return g.zoneId
}
func (g *Gate) ID() string {
	return fmt.Sprintf("%2d%2d", g.ZoneID(), g.id)
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

func New(zoneID, serverId int32) *Gate {
	return &Gate{
		zoneId: zoneID,
		id:     serverId,
	}
}

func init() {
	service.Register(pb.ServiceInfo_Gate, func(info pb.ServiceInfo) service.IService {
		return New(info.Zone, info.ID)
	})
}
