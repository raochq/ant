package service

import (
	"fmt"
	"github.com/raochq/ant/protocol/pb"
)

//ETCD Key
const (
	EKey_Config  = "/Config"  // 服务配置
	EKey_Service = "/Service" // 服务列表
	EKey_Zone    = "/Zone"    // 区服列表
	EKey_State   = "/State"   // 服务状态
	EKey_Addr    = "/Addr"    // 服务端口地址
)

// 服务状态
type State uint8

const (
	Stopped State = iota
	Init
	Running
	Stopping
)

func (s State) String() string {
	switch s {
	case Stopped:
		return "stopped"
	case Init:
		return "init"
	case Running:
		return "running"
	case Stopping:
		return "stopping"
	}
	return fmt.Sprintf("%d", s)
}

// 服务消息类型
type ServerNotify = uint8

const (
	notifyStop ServerNotify = iota
	notifyReloadCSV
	notifyReloadConf
	notifyReport
)

type IService interface {
	Init() error // 初始化
	Stop()       // 停止服务
	Destroy()    // 销毁服务
}

type CreateServiceFunc func(string, pb.ServiceInfo) IService

type Config struct {
	Id       string
	Etcd     []string
	LogPath  string
	LogLevel string
}
