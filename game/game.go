package game

import (
	"github.com/raochq/ant/config"
	"github.com/raochq/ant/protocol/pb"
	"github.com/raochq/ant/service"
)

type Game struct {
	RPCServer
	config *config.Game

	rpcAddr   string
	name      string
	PlayerCnt int32
}
type gameState struct {
	PlayerCnt int32  `json:"cnt"`
	Rpc       string `json:"rpc"`
}

func (g *Game) Name() string {
	return g.name
}

var _ service.IService = (*Game)(nil)

func (g *Game) Init() error {
	cfg := g.config
	addr, err := g.startGrpc(cfg.RPC)
	if err != nil {
		return err
	}
	g.rpcAddr = addr
	return nil
}

func (g *Game) StateInfo() interface{} {
	return gameState{
		PlayerCnt: g.PlayerCnt,
		Rpc:       g.rpcAddr,
	}
}

func (g *Game) Close() {
	g.stopGrpc()
}

func (g *Game) Zone() uint32 {
	return g.config.Zone
}

func New(conf *config.Game) *Game {
	if conf == nil {
		return nil
	}
	g := &Game{
		name:   conf.UniqueID(),
		config: conf,
	}
	g.RPCServer.owner = g
	return g
}

func init() {
	service.Register(pb.ServiceKind_Game.String(), func(conf config.Config) service.IService {
		if c, ok := conf.(*config.Game); ok {
			return New(c)
		}
		return nil
	})
}
