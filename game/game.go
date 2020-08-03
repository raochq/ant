package game

import (
	"context"
	"fmt"
	"github.com/raochq/ant/protocol/pb"
	"github.com/raochq/ant/service"
	"go.etcd.io/etcd/clientv3"
)

type Game struct {
	pb.ServiceInfo
	name string

	state service.State
}

func (g *Game) State() service.State {
	return g.state
}
func (g *Game) Init() error {
	svr, err := service.GetService(g.name)
	if err != nil {
		return err
	}
	svr.WatchETCD(context.TODO(), fmt.Sprintf("%s%s/%d", service.EKey_Service, service.EKey_Zone, svr.Zone), func(evt *clientv3.Event) {
		fmt.Println("watchETCD:", evt)
	})
	return nil
}

func (g *Game) Destroy() {

}
func (g *Game) Stop() {

}

func New(name string, info pb.ServiceInfo) *Game {
	return &Game{
		name:        name,
		ServiceInfo: info,
	}
}

func init() {
	service.Register(pb.ServiceInfo_Game, func(name string, info pb.ServiceInfo) service.IService {
		return New(name, info)
	})
}
