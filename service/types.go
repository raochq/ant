package service

import (
	"fmt"

	"github.com/raochq/ant/protocol/pb"
	clientv3 "go.etcd.io/etcd/client/v3"
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

type IService interface {
	Init() error                                                         // 初始化
	UpdateState(*clientv3.Client, clientv3.LeaseID, string, State) error // 向ETCD更新消息
	Destroy()                                                            // 销毁服务
}

type CreateServiceFunc func(string, pb.ServiceInfo) IService

type Config struct {
	ServiceIDs []string
	Etcd       []string
	LogPath    string
	LogLevel   string
}
