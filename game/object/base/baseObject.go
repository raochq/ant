package base

import (
	"log/slog"
	"time"
)

type ObjectType int32

const (
	OBJ_NONE ObjectType = iota
	OBJ_PLAYER
	OBJ_NPC
	OBJ_Builidng
)

type EventType int32

const (
	Event_Add       EventType = 0
	Event_Remove    EventType = 1
	Event_Move      EventType = 2
	Event_AfterMove EventType = 3
)

type EventHandler func(*Object, any)

type Objecter interface {
	Object() *Object
	CheckMove(Point) bool   // 虚函数，子类可以重写，实现自己的逻辑
	Tick(time.Time)         // 虚函数，子类可以重写，实现自己的逻辑
	ObjectType() ObjectType // 纯虚函数，子类必须实现
}

// ====BaseObject=====
// 公共基类数据和方法
// 涉及多态方法，都不可以直接被调用，必须通过impl调用
type BaseObject struct {
	impl *Object

	// 公共数据
	uuid   int64
	x, y   int32
	omap   *Map
	events map[EventType][]EventHandler
}

func (o *BaseObject) UUID() int64 {
	return o.uuid
}

func (o *BaseObject) Object() *Object {
	return o.impl
}

func (o *BaseObject) Init(impl Objecter, id int64) {
	o.impl = &Object{
		Objecter:   impl,
		BaseObject: o,
	}
	o.uuid = id
	o.events = make(map[EventType][]EventHandler)
}
func (o *BaseObject) SetXY(x, y int32) {
	o.x = x
	o.y = y
}
func (o *BaseObject) GetXY() (int32, int32) {
	return o.x, o.y
}
func (o *BaseObject) OwnMap() *Map {
	return o.omap
}
func (o *BaseObject) SetMap(omap *Map) {
	o.omap = omap
}

func (o *BaseObject) CheckMove(pt Point) bool {
	return false
}

func (o *BaseObject) RegisterEvent(eventType EventType, handler EventHandler) {
	if _, ok := o.events[eventType]; !ok {
		o.events[eventType] = make([]EventHandler, 0)
	}
	o.events[eventType] = append(o.events[eventType], handler)
}

func (o *BaseObject) FireEvent(eventType EventType, target *Object, data any) {
	if handlers, ok := o.events[eventType]; ok {
		for _, handler := range handlers {
			handler(target, data)
		}
	}
}

func (o *BaseObject) Tick(time.Time) {
	o.Print()
}

func (o *BaseObject) Print() {
	slog.Debug("Object tick", "type", o.impl.ObjectType())
}

// =====Object=====
// 对外暴露的Object的对象
// 对外传递的对象都应该使用 *Object
// 非纯虚函数，如果接口的实现类了对应方法，会导致重复定义的编译错误，需要重定向到接口实例的实现
// 纯虚函数，可以不做重定向。
type Object struct {
	Objecter
	*BaseObject
}

func (o *Object) Tick(t time.Time) {
	o.Objecter.Tick(t)
}

// ObjectType是纯虚函数，可以不做重定向
func (o *Object) ObjectType() ObjectType {
	return o.Objecter.ObjectType()
}

func (o *Object) CheckMove(pt Point) bool {
	return o.Objecter.CheckMove(pt)
}
