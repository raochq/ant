package player

import (
	"log/slog"
	"math/rand"
	"time"

	"github.com/raochq/ant/game/object/base"
)

type Player struct {
	base.BaseObject
	Name string
	Age  int
	Sex  bool
}

var _ base.Objecter = (*Player)(nil)

func NewPlayer(id int64, name string, age int, sex bool) (*Player, error) {
	p := &Player{
		Name: name,
		Age:  age,
		Sex:  sex,
	}
	p.Init(p, id)
	return p, nil
}

func (p *Player) ObjectType() base.ObjectType {
	return base.OBJ_PLAYER
}

func (p *Player) Init(impl base.Objecter, id int64) error {
	p.BaseObject.Init(impl, id)
	p.RegisterEvent(base.Event_Add, p.OnAddToMap)
	p.RegisterEvent(base.Event_AfterMove, p.AfterMove)
	return nil
}

func (p *Player) OnAddToMap(target *base.Object, data any) {
	slog.Info("\033[0;34mPlayer\033[0m event added to map", "map id", p.UUID(), "target ID", target.UUID())
}
func (p *Player) AfterMove(target *base.Object, data any) {
	x, y := p.GetXY()
	if target != nil {
		x, y = target.GetXY()
		slog.Info("target is nil")
	}
	slog.Info("\033[0;34mPlayer AfterMove \033[0m", "map id", p.UUID(), "myname", p.Name, slog.Group("pos", "x", x, "y", y))
}

func (p *Player) Tick(t time.Time) {
	// call base.Object's Tick
	p.BaseObject.Tick(t)
	// do something here
	p.Run()
}

func (p *Player) CheckMove(pt base.Point) bool {
	return true
}

func (p *Player) MoveTo(pt base.Point) {
	if m := p.OwnMap(); m != nil && m.MoveObject(p.Object(), pt) {
		p.FireEvent(base.Event_AfterMove, p.Object(), nil)
	}
}

func (p *Player) Run() {
	x, y := p.OwnMap().Range()
	p.MoveTo(base.Point{X: rand.Int31n(x), Y: rand.Int31n(y)})
}
