package config

import (
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/raochq/ant/protocol/pb"
	"github.com/raochq/ant/util/logger"
)

type Game struct {
	Base          //test
	Listen string `yaml:"listen"`
}

func (g *Game) Unmarshal(in []byte) error {
	return yaml.Unmarshal(in, g)
}

func saveDefaultGame(dir string) {
	tp := pb.ServiceKind_Game
	appName := strings.ToLower(tp.String())
	conf := Game{
		Base: Base{
			Type: tp.String(),
			Zone: 1,
			ID:   1,
			DB: DB{
				Addr:   "root:123456@(127.0.0.1:3306)/ant_" + appName + "?timeout=30s&parseTime=true&loc=Local&charset=utf8",
				MaxCon: 10,
			},
			CSV: "./data",
			Redis: Redis{
				Addr:  "127.0.0.1:6379",
				Auth:  "123456",
				Index: 0,
			},
			Log: logInfo{
				Level: 5,
				PATH:  "log/" + appName + ".log",
			},
			ETCD: []string{
				"http://127.0.0.1:2379",
			},
			RPC: "127.0.0.1:0",
		},
		Listen: "127.0.0.1:0",
	}
	appName = filepath.Join(dir, appName+".yml")
	if data, err := yaml.Marshal(conf); err == nil {
		logger.Infof("write default game config file=%v", appName)
		os.WriteFile(appName, data, 0777)
	}
}

func init() {
	register(pb.ServiceKind_Game, func() Config {
		return &Game{}
	})
}
