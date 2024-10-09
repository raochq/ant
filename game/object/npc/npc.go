package npc

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/raochq/ant/game/object/base"
)

type NPC struct {
	*base.BaseObject
	Name string
	Age  int
}

var _ base.Objecter = (*NPC)(nil)

func NewNPC(id int64, name string, age int) (*base.Object, error) {
	npc := &NPC{
		Name: name,
		Age:  age,
	}
	return base.NewObject(npc, id)
}

func (p *NPC) Init(super *base.BaseObject) error {
	slog.Debug("\033[1;31;40mNPC\033[0m Born")
	p.BaseObject = super
	return nil
}

func (p *NPC) ObjectType() base.ObjectType {
	return base.OBJ_NPC
}

func (p *NPC) Tick(t time.Time) {
	p.BaseObject.Tick(t)
	p.SayHello()
	p.Age++
}
func (p *NPC) CheckMove(base.Point) bool {
	return true
}

func (p *NPC) SayHello() {
	fmt.Println("Hello, my name is ", p.Name, " and I am ", p.Age, " years old.")
}
