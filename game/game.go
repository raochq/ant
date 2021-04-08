package game

import (
	"context"
	"fmt"

	"github.com/raochq/ant/protocol/pb"
	"github.com/raochq/ant/service"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type Game struct {
	id     uint32
	zoneID uint32
	config *pb.GameConfig
	RPCServer

	name string
}

var _ service.IService = (*Game)(nil)

func (g *Game) Init() error {
	svr, err := service.GetService(g.name)
	if err != nil {
		return err
	}
	cfg := g.config
	if err = g.startGrpc(cfg.Port); err != nil {
		return err
	}
	svr.WatchETCD(context.TODO(), fmt.Sprintf("%s%s/%d", service.EKey_Service, service.EKey_Zone, svr.Zone), func(evt *clientv3.Event) {
		fmt.Println("watchETCD:", evt)
	})
	return nil
}
func (g *Game) Destroy() {
	g.stopGrpc()
}
func (g *Game) UpdateState(client *clientv3.Client, leaseId clientv3.LeaseID, key string, state service.State) (err error) {
	switch state {
	case service.Running:
		_, err = client.Put(context.TODO(), key+service.EKey_Addr, fmt.Sprintf("%s:%d", g.config.IP, g.config.Port), clientv3.WithLease(leaseId))
	case service.Stopping:
		_, err = client.Delete(context.TODO(), key+service.EKey_Addr)
	}
	return
}
func New(name string, info pb.ServiceInfo) *Game {
	if info.GameConfig == nil {
		return nil
	}
	g := &Game{
		name:   name,
		id:     info.ID,
		zoneID: info.Zone,
		config: info.GameConfig,
	}
	g.RPCServer.owner = g
	return g
}

func init() {
	service.Register(pb.ServiceInfo_Game, func(name string, info pb.ServiceInfo) service.IService {
		return New(name, info)
	})
}
