package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"

	"github.com/raochq/ant/protocol/pb"
)

var confFactory sync.Map

type newConfFn func() Config

type Config interface {
	GetBase() *Base
	Unmarshal([]byte) error
}

func register(kind pb.ServiceKind, fn newConfFn) {
	if fn != nil && kind != pb.ServiceKind_None {
		confFactory.Store(kind, fn)
	}
}

type Base struct {
	Type  string   `yaml:"type"`
	Zone  uint32   `yaml:"zone"`
	ID    uint32   `yaml:"id"`
	CSV   string   `yaml:"csv"`
	DB    DB       `yaml:"db"`
	Redis Redis    `yaml:"redis"`
	Log   logInfo  `yaml:"log"`
	RPC   string   `yaml:"rpc,omitempty"`
	ETCD  []string `yaml:"etcd,flow"`
}

func (c *Base) UniqueID() string {
	return fmt.Sprintf("/Zone/%d/%s/%d", c.Zone, c.Type, c.ID)
}

type Redis struct {
	Timeout int32 `yaml:"timeout"`
	Idle    int32 `yaml:"idle"`

	Addr  string `yaml:"addr"`
	Auth  string `yaml:"auth"`
	Index int32  `yaml:"index"`

	CrosseAddr  string `yaml:"crosseAddr"`
	CrosseAuth  string `yaml:"crosseAuth"`
	CrosseIndex int32  `yaml:"crosseIndex"`
}
type DB struct {
	Addr   string `yaml:"addr"`
	MaxCon int32  `yaml:"maxCon"`
}

type logInfo struct {
	Level int32  `yaml:"level"`
	PATH  string `yaml:"path"`
}

func (c *Base) GetBase() *Base {
	return c
}

func Load(fileName string) (Config, error) {
	buff, err := os.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	node := map[string]interface{}{}
	if err = yaml.Unmarshal(buff, &node); err != nil {
		return nil, err
	}

	typeName := ""
	if base, ok := node["base"]; ok {
		if node, ok = base.(map[string]interface{}); ok {
			if str, o := node["type"]; o {
				typeName, _ = str.(string)
			}
		}
	}
	val, ok := confFactory.Load(pb.ServiceKind(pb.ServiceKind_value[typeName]))
	if !ok {
		return nil, fmt.Errorf("not find config type = %v", typeName)
	}
	newConf := val.(newConfFn)()
	if err = newConf.Unmarshal(buff); err != nil {
		return nil, err
	}
	return newConf, nil
}

func TrySaveDefaultConfig() {
	dir, err := os.Executable()
	if err != nil {
		return
	}
	dir = filepath.Dir(dir)

	saveDefaultLogin(dir)
	saveDefaultGame(dir)
}
