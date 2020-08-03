package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/google/gops/agent"
	"github.com/raochq/ant/engine/logger"
	_ "github.com/raochq/ant/game"
	_ "github.com/raochq/ant/gate"
	"github.com/raochq/ant/service"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var (
	conf         service.Config
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
	flag.StringVar(&confFile, "conf", "", "Location of config, default: $AppName.json")
	flag.BoolVar(&bVersion, "v", false, "Version information")
	flag.Usage = usage
	flag.Parse()
	if bVersion {
		fmt.Print(Version())
		os.Exit(0)
	}
	if len(os.Args) > 1 && confFile == "" {
		flag.Usage()
		os.Exit(0)
	}

	if confFile == "" {
		confFile = strings.TrimSuffix(filepath.Base(os.Args[0]), filepath.Ext(os.Args[0])) + ".json"
	}
	fmt.Printf("conf=%s\n", confFile)
	b, err := ioutil.ReadFile(confFile)
	if err != nil {
		logger.Fatal("load config %v failed %v", confFile, err)
	}

	err = json.Unmarshal(b, &conf)
	if err != nil {
		logger.Fatal("unmarshal config %v failed %v", confFile, err)
	}

	logger.SetOutputFile(conf.LogPath)
	logger.SetLogLevel(conf.LogLevel)
}

func main() {
	Version()
	if err := service.CreateService(conf); err != nil {
		logger.Fatal("register service failed %v", err)
	}

	//使用gops性能监控
	if err := agent.Listen(agent.Options{}); err != nil {
		logger.Fatal("gops listen fail %v", err)
	}
	defer agent.Close()

	service.Run()
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
