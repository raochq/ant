package service

import (
	"fmt"

	"github.com/raochq/ant/config"
)

// ETCD Key
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
	SSStopped State = iota
	SSInit
	SSRunning
	SSStopping
)

func (s State) String() string {
	switch s {
	case SSStopped:
		return "stopped"
	case SSInit:
		return "init"
	case SSRunning:
		return "running"
	case SSStopping:
		return "stopping"
	}
	return fmt.Sprintf("%d", s)
}

type IService interface {
	Init() error            // 初始化
	Close()                 // 关闭服务
	StateInfo() interface{} // 状态信息
}

type CreateServiceFunc func(config.Config) IService

type Config struct {
	ServiceIDs []string
	Etcd       []string
	LogPath    string
	LogLevel   string
}
