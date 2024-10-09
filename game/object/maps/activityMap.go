package maps

import (
	"errors"
	"log/slog"

	"github.com/raochq/ant/game/object/base"
)

type ActivityMap struct {
	*base.BaseMap
	activityOpen bool
}

func (m *ActivityMap) Init(super *base.BaseMap) error {
	m.BaseMap = super
	return nil
}
func (m *ActivityMap) CheckVaildXY(pt base.Point) bool {
	// 只能在活动开启时才能使用
	return m.activityOpen && m.BaseMap.CheckVaildXY(pt)
}

func (m *ActivityMap) AddObject(obj *base.Object) error {
	if m.activityOpen {
		return m.BaseMap.AddObject(obj)
	}
	slog.Error("活动未开启")
	return errors.New("活动未开启")
}

func (m *ActivityMap) MoveObject(o *base.Object, pt base.Point) bool {
	if m.activityOpen {
		slog.Error("活动进行中......")
		return m.BaseMap.MoveObject(o, pt)
	}
	slog.Error("活动未开启")
	return false
}

func NewActivityMap(name string, enabled bool) (*base.Map, error) {
	m := &ActivityMap{
		activityOpen: enabled,
	}

	return base.NewMap(m, name)
}
