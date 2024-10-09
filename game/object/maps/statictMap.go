package maps

import (
	"github.com/raochq/ant/game/object/base"
)

type StaticMap struct {
	*base.BaseMap
	blokcs map[base.Point]struct{}
}

func NewStaticMap(name string) (*base.Map, error) {
	m := &StaticMap{
		blokcs: make(map[base.Point]struct{}),
	}
	return base.NewMap(m, name)
}

func (m *StaticMap) Init(super *base.BaseMap) error {
	m.BaseMap = super
	return nil
}

func (m *StaticMap) CheckVaildXY(pt base.Point) bool {
	if _, ok := m.blokcs[pt]; !ok {
		return m.BaseMap.CheckVaildXY(pt)
	}
	return false
}
