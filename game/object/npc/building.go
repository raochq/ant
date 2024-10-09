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

func NewBuilding(id int64, name string) (*base.Object, error) {
	impl := &Building{
		NPC: NPC{
			Name: name,
		},
	}
	return base.NewObject(impl, id)
}

func (b *Building) ObjectType() base.ObjectType {
	return base.OBJ_Builidng
}

func (b *Building) Init(super *base.BaseObject) error {
	b.NPC.Init(super)
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
