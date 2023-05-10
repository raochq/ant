package service

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/raochq/ant/util/logger"
)

// Service 服务基类
type Service struct {
	IService
	State State
	etcd  *ETCDClient

	name   string
	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
}
type serviceInfo struct {
	Name  string
	State State
	Info  interface{} `json:"info,omitempty"`
}

func (s *Service) Info() string {
	rpcInfo := serviceInfo{
		Name:  s.Name(),
		State: s.State,
		Info:  s.StateInfo(),
	}

	buff, err := json.Marshal(rpcInfo)
	if err != nil {
		return ""
	}
	return string(buff)
}
func (s *Service) Name() string {
	return s.name
}

func (s *Service) Init() error {
	if err := s.IService.Init(); err != nil {
		return err
	}
	logger.WithField("name", s.name).Info("=== Create service ===")
	s.wg.Add(1)
	if err := s.etcd.Start(s.Info(), func() { s.wg.Done() }); err != nil {
		s.wg.Done()
		s.IService.Close()
		return err
	}
	s.SetState(SSRunning)
	logger.WithField("name", s.name).Info("=== running service ===")
	go s.loop()
	return nil
}

//// 监控Key
//func (s *Service) WatchETCD(ctx context.Context, key string, f func(evt *clientv3.Event)) error {
//	if s.client == nil {
//		return errors.New("etcd client is nil")
//	}
//	rch := s.client.Watch(ctx, key, clientv3.WithPrefix())
//	go func() {
//		for {
//			select {
//			case <-ctx.Done():
//				return
//			case wresp := <-rch:
//				for _, ev := range wresp.Events {
//					logger.RPCInfo("%s,find etcd event [%s] %q : %q\n", s.Name(), ev.Type, ev.Kv.Key, ev.Kv.Value)
//					if f != nil {
//						f(ev)
//					}
//				}
//			}
//		}
//	}()
//
//	return nil
//}

func (s *Service) loop() {
	s.wg.Add(1)
	defer s.wg.Done()

	t := time.NewTicker(time.Second * 10)
	defer t.Stop()
	for {
		select {
		case <-s.ctx.Done():
			logger.WithField("name", s.Name()).Info("service loop exit")
			return
		case <-t.C:
			s.UpdateInfo()
		}
	}
}

// Close 关闭服务器
func (s *Service) Close() {
	s.SetState(SSStopping)
	logger.Infof("=== stopping service %v ===", s.Name())
	s.IService.Close()
	s.cancel()
	s.wg.Wait()
	s.State = SSStopped
	logger.Infof("=== stopped service %v ===", s.Name())
}

func (s *Service) SetState(state State) {
	s.State = state
	s.UpdateInfo()
}

// UpdateInfo 更新服务器数据信息,用于数据监控
func (s *Service) UpdateInfo() {
	s.etcd.Update(s.Info())
}
