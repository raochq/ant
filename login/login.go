package login

import (
	"log/slog"
	"net/http"

	"github.com/raochq/ant/config"
	"github.com/raochq/ant/login/dao"
	"github.com/raochq/ant/login/httpService"
	"github.com/raochq/ant/protocol/pb"
	"github.com/raochq/ant/service"
)

type Login struct {
	id     uint32
	zoneID uint32
	config *config.Login
	name   string
	srv    *http.Server
}
type logInfo struct {
	PlayerCnt int32  `json:"cnt"`
	Addr      string `json:"addr"`
}

var _ service.IService = (*Login)(nil)

func (login *Login) StateInfo() interface{} {
	if login.srv != nil {
		return logInfo{
			Addr: login.srv.Addr,
		}
	}
	return nil
}

func (login *Login) Name() string {
	return login.name
}

func (login *Login) Init() error {
	cfg := login.config
	err := dao.InitMysql(cfg.DB.Addr, cfg.DB.MaxCon)
	if err != nil {
		slog.Error("Login init DB failed", "error", err)
		return err
	}
	dao.InitRedis(cfg.Redis)
	login.srv, err = httpService.StartHttpService(cfg.Listen, login.zoneID)
	if err != nil {
		slog.Error("Login startHttpService failed", "error", err)
		return err
	}
	slog.Info("Login init success")
	return nil
}
func (login *Login) Close() {
	if login.srv != nil {
		login.srv.Close()
	}
	slog.Info("Login destroy...")
}

func New(conf *config.Login) *Login {
	if conf == nil {
		return nil
	}
	g := &Login{
		name:   conf.UniqueID(),
		id:     conf.ID,
		zoneID: conf.Zone,
		config: conf,
	}
	return g
}
func init() {
	service.Register(pb.ServiceKind_Login.String(), func(conf config.Config) service.IService {
		if c, ok := conf.(*config.Login); ok {
			return New(c)
		}
		return nil
	})
}
