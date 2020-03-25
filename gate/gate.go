package gate

import (
	"fmt"
	"github.com/raochq/ant/engine/logger"
)

type Gate struct {
	zoneId uint16
	id     uint16
}

func (g *Gate) Name() string {
	return fmt.Sprintf("/Gate/%d/%d", g.zoneId, g.id)
}
func (g *Gate) ZoneID() uint16 {
	return g.zoneId
}
func (g *Gate) ID() uint16 {
	return g.id
}

func (g *Gate) Init() {
	logger.Info("Gate init...\n")
}
func (g *Gate) Destroy() {
	logger.Info("Gate destroy...\n")
}
func (g *Gate) MainLoop(sig <-chan byte) {
	logger.Info("Gate Run in Loop\n")
}

func New(zoneID, serverId uint16) *Gate {
	return &Gate{
		zoneId: zoneID,
		id:     serverId,
	}
}
