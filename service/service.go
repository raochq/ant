package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/raochq/ant/engine/utils"
	"github.com/raochq/ant/protocol/pb"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/raochq/ant/engine/logger"
	"go.etcd.io/etcd/clientv3"
)

var (
	gServiceFactory sync.Map // map[pb.ServiceInfo_SKind]reflect.Type{}
	gServiceList    sync.Map // map[string]*Service{}
)

// 注册服务
func Register(k pb.ServiceInfo_SKind, fn CreateServiceFunc) {
	if _, ok := gServiceFactory.LoadOrStore(k, fn); ok {
		logger.Fatal("repeated register service %v", k)
	}
}

//服务基类
type Service struct {
	IService
	pb.ServiceInfo

	name   string
	chStop chan ServerNotify
	state  State

	leaseId  clientv3.LeaseID
	client   *clientv3.Client
	leaseSig <-chan *clientv3.LeaseKeepAliveResponse
}

// 从etcd查找服务二配置
func getServiceConf(Id string, address []string) (*pb.ServiceInfo, error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   address,
		DialTimeout: time.Second,
	})
	if err != nil {
		logger.Fatal("connect etcd failed %v", err)
	}
	defer cli.Close()
	key := EKey_Config + Id
	resp, err := cli.Get(context.Background(), key)
	if err != nil {
		return nil, err
	}
	kvs := resp.Kvs
	if len(kvs) != 1 {
		return nil, fmt.Errorf("invalid response for key: %s", key)
	}
	kv := kvs[0]

	var info pb.ServiceInfo
	err = json.Unmarshal(kv.Value, &info)
	if err != nil {
		return nil, err
	}
	return &info, nil
}

// 创建服务
func CreateService(conf Config) error {
	info, err := getServiceConf(conf.Id, conf.Etcd)
	if err != nil {
		return fmt.Errorf("GetService error %w", err)
	}
	obj, ok := gServiceFactory.Load(info.Kind)
	if !ok {
		return fmt.Errorf("unknown service %v", info.Kind)
	}
	fn := obj.(CreateServiceFunc)

	svr := &Service{
		IService:    fn(conf.Id, *info),
		ServiceInfo: *info,
		name:        conf.Id,
		chStop:      make(chan ServerNotify),
	}
	if _, ok = gServiceList.LoadOrStore(svr.name, svr); ok {
		return fmt.Errorf("service %s is already registered", svr.name)
	}
	err = svr.registerToETCD(conf.Etcd)
	if err != nil {
		gServiceList.Delete(svr.name)
		return err
	}
	logger.Info("===Create service %v ===", svr.name)
	return nil
}

//服务启动
func Run() {
	var wg sync.WaitGroup
	//启动服务
	gServiceList.Range(func(_, value interface{}) bool {
		s := value.(*Service)
		wg.Add(1)
		go s.run(&wg)
		return true
	})
	WaitForStop()
	// 退出服务
	gServiceList.Range(func(_, value interface{}) bool {
		s := value.(*Service)
		s.close()
		return true
	})
	wg.Wait()
}
func WaitForStop() {
	//监听外部信号
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGABRT, syscall.SIGTERM)
	sig := <-ch
	logger.Debug("捕获%v（%v）信号， 退出服务", sig, int(sig.(syscall.Signal)))
}

// 获取服务实例
func GetService(name string) (*Service, error) {
	if value, ok := gServiceList.Load(name); ok {
		return value.(*Service), nil
	}
	return nil, fmt.Errorf("service %s not found", name)
}

//计算服务唯一Key
func GetKey(info *pb.ServiceInfo) string {
	if info == nil {
		return ""
	}
	return fmt.Sprintf("%s/%d/%s/%d", EKey_Zone, info.Zone, info.Kind, info.ID)
}

// 向ETCD请求注册
func (s *Service) registerToETCD(address []string) error {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   address,
		DialTimeout: time.Second,
	})
	if err != nil {
		logger.Fatal("Create etcd client failed", err)
	}

	resp, err := cli.Grant(context.TODO(), 5)
	if err != nil {
		logger.Fatal("connect to etcd failed %v", err)
	}
	s.client = cli
	s.leaseId = resp.ID
	s.state = Init

	err = s.synState()
	if err != nil {
		logger.Fatal("put to etcd failed %v", err)
	}

	s.leaseSig, err = s.client.KeepAlive(context.TODO(), s.leaseId)
	if err != nil {
		logger.Fatal("KeepAlive to etcd failed %v", err)
	}
	return nil
}

//监控Key
func (s *Service) WatchETCD(ctx context.Context, key string, f func(evt *clientv3.Event)) error {
	if s.client == nil {
		return errors.New("etcd client is nil")
	}
	rch := s.client.Watch(ctx, key, clientv3.WithPrefix())
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case wresp := <-rch:
				for _, ev := range wresp.Events {
					logger.Info("%s,find etcd event [%s] %q : %q\n", s.name, ev.Type, ev.Kv.Key, ev.Kv.Value)
					if f != nil {
						f(ev)
					}
				}
			}
		}
	}()

	return nil
}

func (s *Service) synState() (err error) {
	if s.client == nil {
		return errors.New("client is nil")
	}
	switch s.state {
	case Init:
		_, err = s.client.Put(context.TODO(), EKey_Service+s.Key()+EKey_State, s.state.String(), clientv3.WithLease(s.leaseId))
		if err != nil {
			logger.Error("%s synState error: %v", s.Key(), err)
			return
		}
	case Running:
		_, err = s.client.Put(context.TODO(), EKey_Service+s.Key()+EKey_Addr, fmt.Sprintf("%s:%d", s.IP, s.Port), clientv3.WithLease(s.leaseId))
		if err != nil {
			logger.Error("%s synState error: %v", s.Key(), err)
			return
		}
		_, err = s.client.Put(context.TODO(), EKey_Service+s.Key()+EKey_State, s.state.String(), clientv3.WithLease(s.leaseId))
		if err != nil {
			logger.Error("%s synState error: %v", s.Key(), err)
			return
		}
	case Stopping:
		_, err = s.client.Put(context.TODO(), EKey_Service+s.Key()+EKey_State, s.state.String(), clientv3.WithLease(s.leaseId))
		if err != nil {
			logger.Error("%s synState error: %v", s.Key(), err)
			return
		}
		_, err = s.client.Delete(context.TODO(), EKey_Service+s.Key()+EKey_Addr)
		if err != nil {
			logger.Error("%s synState error: %v", s.Key(), err)
			return
		}
	case Stopped:
		_, err = s.client.Delete(context.TODO(), EKey_Service+s.Key()+EKey_State)
		if err != nil {
			logger.Error("%s synState error: %v", s.Key(), err)
			return
		}
	}
	logger.Info("%s synState: %v", s.Key(), s.state)
	return err
}
func (s *Service) run(w *sync.WaitGroup) {
	defer w.Done()
	logger.Info("===init service %v ===", s.name)
	s.Init()
	s.state = Running
	logger.Info("===running service %v ===", s.name)
	s.synState()
	s.loop()
	s.Destroy()
	s.state = Stopped
	s.synState()
	s.client.Revoke(context.TODO(), s.leaseId)
	s.leaseId = 0
	logger.Info("===stopped service %v ===", s.name)

}
func (s *Service) loop() {
	defer utils.PrintPanicStack()
	defer s.stop()

	for {
		select {
		case c, ok := <-s.chStop:
			if !ok {
				return
			}
			switch c {
			case notifyStop:
				return
			case notifyReloadCSV:
			case notifyReloadConf:
			case notifyReport:
			default:
				return
			}
		case <-s.leaseSig:
		}
	}

}

func (s *Service) close() {
	s.chStop <- notifyStop
}
func (s *Service) stop() {
	s.state = Stopping
	s.synState()
	logger.Info("===stopping service %v ===", s.name)
	s.Stop()
}
func (s *Service) Key() string {
	return GetKey(&s.ServiceInfo)
}
