package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/google/gops/agent"
	"github.com/raochq/ant/engine/logger"
	"github.com/raochq/ant/game"
	"github.com/raochq/ant/gate"
	"github.com/raochq/ant/service"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

var (
	zoneID       uint
	serverID     uint
	etcd         []string
	ServiceName  string
	conf         string
	buildVersion = BuildVersion{"2020-03-19", "1.0.1", ""}
)

type BuildVersion struct {
	BuildData string
	BuildVer  string
	GitTag    string
}

func init() {
	var etcdstr string
	h := flag.Bool("h", false, "help    :显示帮助")
	v := flag.Bool("v", false, "version :查看版本信息")
	flag.StringVar(&ServiceName, "s", "", "service :需要启动的服务模块名称")
	flag.StringVar(&conf, "c", "", "conf :配置文件位置")
	flag.StringVar(&etcdstr, "e", "127.0.0.1:2379", "etcd :etcd地址: etcd1,etcd2,etcd3...")
	flag.UintVar(&zoneID, "zid", 1, "zoneID :区服id")
	flag.UintVar(&zoneID, "sid", 1, "zoneID :服务器id")
	flag.Usage = usage
	flag.Parse()
	if *v {
		fmt.Printf("%v\n", buildVersion)
		os.Exit(0)
	}

	if *h || ServiceName == "" {
		flag.Usage()
		os.Exit(0)
	}
	etcd = strings.Split(etcdstr, ",")

	loadConfig()
}

func loadConfig() {
	if conf == "" {
		return
	}
	b, err := ioutil.ReadFile(conf)
	if err != nil {
		logger.Fatal("load config %v failed %v", conf, err)
	}

	cfg := struct {
		Etcd    []string
		LogFile string
	}{}
	err = json.Unmarshal(b, &cfg)
	if err != nil {
		logger.Fatal("unmarshal config %v failed %v", conf, err)
	}
	logger.SetOutput(cfg.LogFile)
	etcd = cfg.Etcd
}

func main() {
	monitor()

	var svr service.IService
	switch ServiceName {
	case "game":
		svr = game.New(uint16(zoneID), uint16(serverID))
	case "gate":
		svr = gate.New(uint16(zoneID), uint16(serverID))
	default:
		log.Fatal("unknown service name ", ServiceName)
	}
	service.Register(svr, etcd)
	service.Run()
}

func usage() {
	const helpText = `Usage:
  ant <cmd> [options]
options:
`
	fmt.Fprintf(os.Stderr, "ant version: %s\n%s", buildVersion.BuildVer, helpText)
	flag.PrintDefaults()
}

//使用gops性能监控
func monitor() {
	if err := agent.Listen(agent.Options{
		Addr:            "",
		ConfigDir:       "",
		ShutdownCleanup: true,
	}); err != nil {
		logger.Fatal("gops listen fail %v", err)
	}
}
