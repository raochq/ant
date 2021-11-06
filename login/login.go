package login

import (
	"context"
	"fmt"
	"net/http"

	"github.com/raochq/ant/engine/logger"
	"github.com/raochq/ant/login/dao"
	"github.com/raochq/ant/login/httpService"
	"github.com/raochq/ant/protocol/pb"
	"github.com/raochq/ant/service"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type Login struct {
	id     uint32
	zoneID uint32
	config *pb.LoginConfig
	name   string
	srv    *http.Server
}

var _ service.IService = (*Login)(nil)

func (login *Login) Init() error {
	cfg := login.config
	err := dao.InitMysql(cfg.DbURL, cfg.DbMaxCon)
	if err != nil {
		logger.Error("Login init DB failed: %v", err)
		return err
	}
	dao.InitRedis(cfg)
	login.srv, err = httpService.StartHttpService(cfg.Port, login.zoneID)
	if err != nil {
		logger.Error("Login startHttpService failed: %v", err)
		return err
	}
	logger.Info("Login init success")
	return nil
}
func (login *Login) Destroy() {
	login.srv.Close()
	logger.Info("Login destroy...\n")
}

func (login *Login) UpdateState(client *clientv3.Client, leaseId clientv3.LeaseID, key string, state service.State) (err error) {
	switch state {
	case service.Running:
		_, err = client.Put(context.TODO(), key+service.EKey_Addr, fmt.Sprintf("%s:%d", login.config.IP, login.config.Port), clientv3.WithLease(leaseId))
	case service.Stopping:
		_, err = client.Delete(context.TODO(), key+service.EKey_Addr)
	}
	return
}
func New(name string, info pb.ServiceInfo) *Login {
	if info.LoginConfig == nil {
		return nil
	}
	g := &Login{
		name:   name,
		id:     info.ID,
		zoneID: info.Zone,
		config: info.LoginConfig,
	}
	return g
}
func init() {
	service.Register(pb.ServiceInfo_Login, func(name string, info pb.ServiceInfo) service.IService {
		return New(name, info)
	})
}
