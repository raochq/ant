package game

import "fmt"

type Game struct {
	zoneId uint16
	id     uint16
}

func (g *Game) Name() string {
	return fmt.Sprintf("Game/%d/%d", g.zoneId, g.id)
}
func (g *Game) ZoneID() uint16 {
	return g.zoneId
}
func (g *Game) ID() uint16 {
	return g.id
}
func (g *Game) Init() {

}
func (g *Game) Destroy() {

}
func (g *Game) MainLoop(sig <-chan byte) {

}

func New(zoneID, serverId uint16) *Game {
	return &Game{
		zoneId: zoneID,
		id:     serverId,
	}
}
