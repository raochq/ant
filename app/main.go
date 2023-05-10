package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/google/gops/agent"

	"github.com/raochq/ant/config"
	_ "github.com/raochq/ant/game"
	_ "github.com/raochq/ant/gate"
	_ "github.com/raochq/ant/login"
	"github.com/raochq/ant/service"
	"github.com/raochq/ant/util/logger"
)

var (
	confs        []config.Config
	AppName      = "ant" // 应用名称
	AppVersion   string  // 应用版本
	BuildVersion string  // 编译版本
	BuildTime    string  // 编译时间
	GitRevision  string  // Git版本
	GitBranch    string  // Git分支
	GoVersion    string  // Golang信息
)

func init() {
	var confFile string
	bVersion := false
	initConf := false
	flag.StringVar(&confFile, "conf", "", "service of config,split by comma(,), default: all yaml file in current directory")
	flag.BoolVar(&bVersion, "v", false, "Version information")
	flag.BoolVar(&initConf, "init", false, "create an default config")
	flag.Usage = usage
	flag.Parse()
	if bVersion {
		fmt.Print(Version())
		os.Exit(0)
	}

	if initConf {
		config.TrySaveDefaultConfig()
		os.Exit(0)
	}
	if len(os.Args) > 1 && confFile == "" {
		flag.Usage()
		os.Exit(0)
	}
	var list []string
	if confFile != "" {
		list = strings.Split(confFile, ",")
	} else {
		list = allConfig()
	}

	logger.Infof("conf=%s\n", confFile)
	for _, c := range list {
		conf, err := config.Load(c)
		if err != nil {
			logger.Fatalf("load config %v failed %v", confFile, err)
		}
		confs = append(confs, conf)
	}
}
func allConfig() []string {
	curr := "./bin"
	dir, err := os.ReadDir(curr)
	if err != nil {
		return nil
	}
	var ans []string
	for _, fi := range dir {
		if fi.IsDir() {
			continue
		}
		name := strings.ToLower(fi.Name())
		if strings.HasSuffix(name, ".yml") || strings.HasSuffix(name, ".yaml") {
			ans = append(ans, filepath.Join(curr, fi.Name()))
		}
	}
	return ans
}

func main() {
	logger.Info(Version())
	if err := service.StartService(confs); err != nil {
		logger.Fatal("register service failed %v", err)
	}
	//使用gops性能监控
	if err := agent.Listen(agent.Options{}); err != nil {
		logger.Fatal("gops listen fail %v", err)
	}
	defer agent.Close()

	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGKILL, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGABRT, syscall.SIGTERM)
	sig := <-ch
	logger.WithField("sig", fmt.Sprintf("%v(%d)", sig, int(sig.(syscall.Signal)))).Info("capture signal, exit service")

	service.CloseAll()
	logger.Info("=== close All service ===")
}

func usage() {
	const helpText = `Usage:
  ant <cmd> [options]
options:
`
	fmt.Fprintf(os.Stderr, "%s\n%s", Version(), helpText)
	flag.PrintDefaults()
}

// Version 版本信息
func Version() string {
	return fmt.Sprintf("%s Version:\t%s\nBuild version:\t%s\nBuild time:\t%s\nGit revision:\t%s\nGit branch:\t%s\nGolang Version: %s\n", AppName, AppVersion, BuildVersion, BuildTime, GitRevision, GitBranch, GoVersion)
}
