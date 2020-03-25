package service

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/raochq/ant/engine/logger"
	"go.etcd.io/etcd/clientv3"
)

const (
	SIG_STOP   = iota // 停止
	SIG_RELOAD        // 重载配置
	SIG_REPORT        // 上报状态信息
)

var services []*Service

type IService interface {
	Name() string
	MainLoop(sig <-chan byte)
	ZoneID() uint16
	ID() uint16
}

//服务基类
type Service struct {
	impl IService
	sig  chan byte
	wg   sync.WaitGroup
}

func Register(s IService, address []string) {
	if s == nil {
		logger.Fatal("Register a nil IService")
	}
	svr := new(Service)
	svr.sig = make(chan byte)
	svr.impl = s
	if err := svr.registerToETCD(address); err != nil {
		logger.Fatal("%s register to etcd failed %v", s.Name(), err)
	}

	services = append(services, svr)
	logger.Info("registered service %v ok", s.Name())
}

func onInit() {
	for _, s := range services {
		s.wg.Add(1)
		go func() {
			s.impl.MainLoop(s.sig)
			s.wg.Done()
		}()
	}
}

func stop() {
	for i := len(services) - 1; i >= 0; i-- {
		s := services[i]
		s.sig <- SIG_STOP
		s.wg.Wait()
	}
}
func Reload() {
	for _, s := range services {
		select {
		case s.sig <- SIG_RELOAD:
		default:
			logger.Error("Reload failed for %v", s.impl.Name())
		}

	}
}
func ReportState() {
	for _, s := range services {
		select {
		case s.sig <- SIG_REPORT:
		default:
			logger.Error("ReportState failed for %v", s.impl.Name())
		}
	}
}

//服务启动
func Run() {
	if len(services) == 0 {
		logger.Fatal("nona services to run!")
	}
	onInit()
	listenSignal()
}

//监听外部信号
func listenSignal() {
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGABRT, syscall.SIGTERM, syscall.SIGHUP, syscall.Signal(0x1e), syscall.Signal(0x1f), syscall.Signal(29))
	for {
		sig := <-ch
		switch sig {
		case syscall.SIGHUP: //重新读配置文件
			logger.Info("获得SIGHUP（1）信号， reload服务")
			Reload()
		case syscall.Signal(29): //获取服务状态
			logger.Info("获得SIGIO（29）信号， 获取服务状态")
			ReportState()
		default: // 退出
			logger.Info("获得%v（%v）信号， 退出服务", sig, int(sig.(syscall.Signal)))
			stop()
			return
		}
	}
}

// 向ETCD请求注册
func (s *Service) registerToETCD(address []string) error {
	cfg := clientv3.Config{
		Endpoints:   address,
		DialTimeout: time.Second,
	}
	cli, err := clientv3.New(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer cli.Close()
	key := s.impl.Name()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)

	_, err = cli.Put(ctx, key, "bbb")
	cancel()
	if err != nil {
		fmt.Println("err:", err)
		return err
	}
	resp, err := cli.Get(context.Background(), key)
	if err != nil {
		fmt.Println("err:", err)
		return err
	}
	for _, kv := range resp.Kvs {
		fmt.Println(string(kv.Key), string(kv.Value))
	}
	return nil
}
