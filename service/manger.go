package service

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/raochq/ant/config"
	"github.com/raochq/ant/util/logger"
)

var mgr = &Manger{
	factory: map[string]CreateServiceFunc{},
	list:    map[string]*Service{},
}

func Manager() *Manger {
	return mgr
}

type Manger struct {
	sync.Mutex
	factory map[string]CreateServiceFunc
	list    map[string]*Service
}

func (m *Manger) register(key string, value CreateServiceFunc) {
	m.Lock()
	defer m.Unlock()

	if _, ok := m.factory[key]; !ok {
		m.factory[key] = value
	}
}
func (m *Manger) getService(name string) *Service {
	m.Lock()
	defer m.Unlock()
	return m.list[name]
}

func (m *Manger) newServiceImpl(conf config.Config) (*Service, error) {
	base := conf.GetBase()
	fn, ok := m.factory[base.Type]
	if !ok {
		return nil, fmt.Errorf("unknown service %v", base.Type)
	}
	newSvr := fn(conf)
	if newSvr == nil {
		return nil, fmt.Errorf("CreateService %v failed", base.Type)
	}
	svr := &Service{
		IService: newSvr,
		State:    SSInit,
		name:     base.UniqueID(),
	}
	svr.ctx, svr.cancel = context.WithCancel(context.TODO())

	if _, ok := m.list[svr.Name()]; ok {
		err := fmt.Errorf("service %s is already registered", svr.Name())
		logger.Error("Create Service failed", err)
		return nil, err
	}
	m.list[svr.Name()] = svr

	if len(base.ETCD) != 0 {
		etcd, err := NewETCDClient(svr.ctx, EKey_Service+base.UniqueID(), base.ETCD)
		if err != nil {
			delete(m.list, svr.Name())
			return nil, err
		}
		svr.etcd = etcd
	}
	return svr, nil
}

func (m *Manger) createService(conf config.Config) (*Service, error) {
	if conf == nil || conf.GetBase() == nil {
		err := errors.New("invalid config for create service")
		logger.WithError(err).Error("Create Service failed")
		return nil, err
	}
	m.Lock()
	defer m.Unlock()

	svr, err := m.newServiceImpl(conf)
	if err != nil {
		return nil, err
	}

	if err := svr.Init(); err != nil {
		logger.WithError(err).WithField("name", svr.Name()).Error("service init failed")
		delete(m.list, svr.Name())
		return nil, err
	}
	return svr, nil
}

func (m *Manger) close() {
	m.Lock()
	defer m.Unlock()
	var wg sync.WaitGroup
	for i := range m.list {
		svr := m.list[i]
		wg.Add(1)
		go func() {
			defer wg.Done()
			svr.Close()
		}()
	}
	wg.Wait()
}

// Register 注册服务
func Register(k string, fn CreateServiceFunc) {
	Manager().register(k, fn)
}

// StartService 创建服务
func StartService(confs []config.Config) error {
	if len(confs) == 0 {
		return errors.New("no config for service")
	}
	logConf := confs[0].GetBase().Log
	logger.SetOutputFile(logConf.PATH)
	logger.SetLogLevel(logConf.Level)

	for _, conf := range confs {
		_, err := Manager().createService(conf)
		if err != nil {
			CloseAll()
			return err
		}
	}
	return nil
}

// CloseAll 退出服务
func CloseAll() {
	Manager().close()
}

// GetService 获取服务实例
func GetService(name string) *Service {
	return Manager().getService(name)
}
