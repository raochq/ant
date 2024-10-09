package base

import (
	"errors"
	"log/slog"
	"math/rand"
	"slices"
	"time"
)

type Maper interface {
	Init(data *BaseMap) error
	CheckVaildXY(pt Point) bool
	AddObject(obj *Object) error
	MoveObject(o *Object, pt Point) bool
}

type Point struct {
	X, Y int32
}

// map的公共数据和方法
// 涉及多态方法，都不可以直接被调用，必须通过impl调用
type BaseMap struct {
	impl *Map

	// 公共数据
	name       string
	width      int32
	height     int32
	objList    []*Object
	objTypeMap map[ObjectType]map[int64]*Object
}

func (m *BaseMap) init(impl *Map, name string) error {
	m.impl = impl
	m.name = name
	m.objTypeMap = make(map[ObjectType]map[int64]*Object)
	m.height = 100
	m.width = 100
	slog.Debug("init map", "name", m.name)
	return nil
}
func (m *BaseMap) Impl() *Map {
	return m.impl
}

func (m *BaseMap) GetObject(id int64) *Object {
	for _, obj := range m.objList {
		if obj.UUID() == id {
			return obj
		}
	}
	return nil
}

func (m *BaseMap) RemoveObject(obj *Object) {
	id := obj.UUID()
	for i, o := range m.objList {
		if o.UUID() == id {
			m.objList = slices.Delete(m.objList, i, i+1)
			break
		}
	}
	delete(m.objTypeMap[obj.ObjectType()], id)
	m.ForEachAll(func(o *Object) {
		o.FireEvent(Event_Remove, obj, nil)
	})
	obj.omap = nil
}

func (m *BaseMap) ForEach(tp ObjectType, f func(obj *Object)) {
	for _, obj := range m.objTypeMap[tp] {
		f(obj)
	}
}

func (m *BaseMap) ForEachAll(f func(obj *Object)) {
	for _, obj := range m.objList {
		f(obj)
	}
}

func (m *BaseMap) Tick(t time.Time) {
	slog.Info("tick", "map", m.name, "time", t)
	m.ForEachAll(func(obj *Object) {
		obj.Tick(t)
	})
}
func (m *BaseMap) Range() (int32, int32) {
	return m.width, m.height
}

func (m *BaseMap) CheckVaildXY(pt Point) bool {
	return pt.X >= 0 && pt.X < m.width && pt.Y >= 0 && pt.Y < m.height
}

func (m *BaseMap) MoveObject(o *Object, pt Point) bool {
	if !m.impl.CheckVaildXY(pt) {
		return false
	}

	if o.CheckMove(pt) {
		o.SetXY(pt.X, pt.Y)
		return true
	}
	return false
}

func (m *BaseMap) AddObject(obj *Object) error {
	id := obj.UUID()
	tp := obj.ObjectType()
	if _, ok := m.objTypeMap[tp][id]; ok {
		return errors.New("object already exists")
	}
	m.objList = append(m.objList, obj)
	ty := obj.ObjectType()
	v, ok := m.objTypeMap[tp]
	if !ok {
		v = make(map[int64]*Object)
		m.objTypeMap[ty] = v
	}
	v[id] = obj
	obj.omap = m.Impl()
	obj.SetXY(rand.Int31n(1000), rand.Int31n(1000))
	m.ForEachAll(func(o *Object) {
		o.FireEvent(Event_Add, obj, nil)
	})
	return nil
}

// ============= Map =============
// 对外暴露的Map对象，对Maper接口的实现
// 对外传递的对象都应该使用*Map
// 非纯虚函数，如果接口的实现类了对应方法，会导致重复定义的编译错误，需要重定向到接口实例的实现
// 纯虚函数，可以不做重定向。
type Map struct {
	BaseMap
	Maper
}

func NewMap(impl Maper, name string) (*Map, error) {
	if impl == nil {
		return nil, errors.New("invalid param")
	}

	if _, ok := impl.(*Map); ok {
		// Map不能作为impl,否则会造成死循环
		return nil, errors.New("can not use map as impl")
	}

	m := &Map{
		Maper: impl,
	}
	m.init(m, name)

	if err := m.Init(&m.BaseMap); err != nil {
		return nil, err
	}
	return m, nil
}

func (m *Map) AddObject(obj *Object) error {
	return m.Maper.AddObject(obj)
}

func (m *Map) MoveObject(o *Object, pt Point) bool {
	return m.Maper.MoveObject(o, pt)
}

func (m *Map) CheckVaildXY(pt Point) bool {
	return m.Maper.CheckVaildXY(pt)
}
