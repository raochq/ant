package game

import (
	"fmt"
	"github.com/raochq/ant/protocol/pb"
	"github.com/raochq/ant/service"
)

type Game struct {
	zoneId int32
	id     int32
	state  service.State
}

func (g *Game) Name() string {
	return fmt.Sprintf("/Game/%d/%d", g.zoneId, g.id)
}
func (g *Game) ZoneID() int32 {
	return g.zoneId
}
func (g *Game) ID() string {
	return fmt.Sprintf("%2d%2d", g.ZoneID(), g.id)
}
func (g *Game) State() service.State {
	return g.state
}

func (g *Game) Init() error {
	return nil
}

func (g *Game) Destroy() {

}
func (g *Game) Stop() {

}

func New(zoneID, serverId int32) *Game {
	return &Game{
		zoneId: zoneID,
		id:     serverId,
	}
}

func init() {
	service.Register(pb.ServiceInfo_Game, func(info pb.ServiceInfo) service.IService {
		return New(info.Zone, info.ID)
	})
}
