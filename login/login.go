package login

import (
	"net/http"

	"github.com/raochq/ant/config"
	"github.com/raochq/ant/login/dao"
	"github.com/raochq/ant/login/httpService"
	"github.com/raochq/ant/protocol/pb"
	"github.com/raochq/ant/service"
	"github.com/raochq/ant/util/logger"
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
		logger.Error("Login init DB failed: %v", err)
		return err
	}
	dao.InitRedis(cfg.Redis)
	login.srv, err = httpService.StartHttpService(cfg.Listen, login.zoneID)
	if err != nil {
		logger.Error("Login startHttpService failed: %v", err)
		return err
	}
	logger.Info("Login init success")
	return nil
}
func (login *Login) Close() {
	login.srv.Close()
	logger.Info("Login destroy...\n")
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
