package npc

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/raochq/ant/game/object/base"
)

type NPC struct {
	base.BaseObject
	Name string
	Age  int
}

var _ base.Objecter = (*NPC)(nil)

func NewNPC(id int64, name string, age int) (*NPC, error) {
	npc := &NPC{
		Name: name,
		Age:  age,
	}
	npc.Init(npc, id)
	return npc, nil
}

func (p *NPC) Init(impl base.Objecter, id int64) error {
	slog.Debug("\033[1;31;40mNPC\033[0m Init")
	p.BaseObject.Init(impl, id)
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
