package npc

import (
	"log/slog"
	"time"

	"github.com/raochq/ant/game/object/base"
)

type Building struct {
	NPC
	zones map[int64]base.Point
}

func NewBuilding(id int64, name string) (*Building, error) {
	b := &Building{
		NPC: NPC{
			Name: name,
		},
	}
	b.Init(b, id)
	return b, nil
}

func (b *Building) ObjectType() base.ObjectType {
	return base.OBJ_Builidng
}

func (b *Building) Init(impl base.Objecter, id int64) error {
	slog.Debug("\033[1;31;40mBuilding\033[0m Init")
	b.NPC.Init(impl, id)
	b.zones = make(map[int64]base.Point)
	return nil
}
func (p *Building) CheckMove(base.Point) bool {
	return false
}
func (b *Building) Tick(t time.Time) {
	b.NPC.Tick(t)
	slog.Info("Building tick.......")
}
