package gate

import (
	"context"
	"fmt"

	"github.com/raochq/ant/engine/logger"
	"github.com/raochq/ant/protocol/pb"
	"github.com/raochq/ant/service"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type Gate struct {
	id     uint32
	zoneID uint32
	config *pb.GateConfig
	name   string

	state service.State
}

func (g *Gate) Stop() {
}

func (g *Gate) Init() error {
	logger.Info("Gate init...\n")
	return nil
}
func (g *Gate) Destroy() {
	logger.Info("Gate destroy...\n")
}
func (g *Gate) UpdateState(client *clientv3.Client, leaseId clientv3.LeaseID, key string, state service.State) (err error) {
	switch state {
	case service.Running:
		_, err = client.Put(context.TODO(), key+service.EKey_Addr, fmt.Sprintf("%s:%d", g.config.IP, g.config.Port), clientv3.WithLease(leaseId))
	case service.Stopping:
		_, err = client.Delete(context.TODO(), key+service.EKey_Addr)
	}
	return nil
}
func (g *Gate) MainLoop(sig <-chan byte) {
	logger.Info("Gate Run in Loop\n")
}

func New(name string, info pb.ServiceInfo) *Gate {
	if info.GateConfig == nil {
		return nil
	}
	return &Gate{
		name:   name,
		id:     info.ID,
		zoneID: info.Zone,
		config: info.GateConfig,
	}
}
func init() {
	service.Register(pb.ServiceInfo_Gate, func(name string, info pb.ServiceInfo) service.IService {
		return New(name, info)
	})
}
