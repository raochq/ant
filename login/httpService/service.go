package httpService

import (
	"fmt"
	"log/slog"
	"net/url"
	"strconv"
	"time"

	"github.com/raochq/ant/common"
	"github.com/raochq/ant/login/dao"
	"github.com/raochq/ant/protocol/pb"
	"github.com/raochq/ant/util"
)

type LoginService struct {
	zoneID uint32
}

func NewLoginService(zoneID uint32) *LoginService {
	return &LoginService{
		zoneID: zoneID,
	}
}

// 获取唯一角色id
func (srv *LoginService) getUniqueAccountID() (int64, error) {
	// todo: id号上限检查
	ticket64, err := dao.GetTicket64(srv.zoneID)
	if err != nil {
		slog.Error("GetUniqueAccountID: redis.T64.GetTicket64 failed", "zoneid", srv.zoneID, "error", err)
		return 0, common.RC_Unknown
	}
	ticket64, err = dao.GenerateTicketPoolID(srv.zoneID)
	if err != nil {
		slog.Error("GetUniqueAccountID: redis.T64.GenerateTicketPoolID failed", srv.zoneID, "error", err)
		return 0, common.RC_Unknown
	}
	// todo: id号上限检查
	return ticket64, nil
}

// 用户登录
func (srv *LoginService) userLogin(account *pb.Account) error {
	//todo: 登录限制检查
	accountDB, err := dao.FindOneByUserNameForUpdate(account.UserName)
	if err != nil {
		slog.Error("userDao.FindOneByUserNameForUpdate failed", slog.Group("account", "UserName", account.UserName, "PassHash", account.PassHash, "LastIP", account.LastIP), "error", err)
		return common.RC_AccoutNotExists
	}
	if accountDB == nil {
		return common.RC_AccoutNotExists
	}
	// verify the password
	if account.PassHash != accountDB.PassHash {
		slog.Error("service.Login, password not match", slog.Group("account", "UserName", account.UserName, "PassHash", account.PassHash))
		return common.RC_PasswordInvalid
	}

	return nil
}

// 注册账号
func (srv *LoginService) doRegisterAccount(openID string, token string, ip string, platform int) (*pb.Account, error) {
	n := time.Now().Unix()
	account := &pb.Account{}
	account.UserName = openID //surprise :-)
	account.PassHash = util.Sha1Hash(openID + token)
	account.LastIP = ip
	account.Platform = uint32(platform)
	account.UserToken = util.CalculateUserToken(account.PassHash)
	account.CTime = n
	account.MTime = n
	account.IsLoggedIn = true
	//todo:渠道注册人数检测

	ticket64, ret := srv.getUniqueAccountID()
	if ret != nil {
		slog.Error("RegisterAccount service.GetUniqueAccountID() failed",
			"openID", openID, "PassHash", account.PassHash, "ip", ip, "error", ret)
		return nil, ret
	}
	slog.Debug(fmt.Sprintf("ticket64: %v", ticket64))
	account.ID = ticket64
	_, err := dao.AddAccount(account)
	if err != nil {
		slog.Error("userDao.AddAccount failed", "openID", openID, "PassHash", account.PassHash, "ip", ip, "error", err)
		return nil, common.RC_RegisterFailure
	}
	return account, nil
}
func (srv *LoginService) registerAccount(openID string, token string, ip string, platform int) (*pb.Account, error) {
	slog.Info("register", "username", openID, "lastIP", ip)
	account, err := dao.FindOneByUserNameForUpdate(openID)
	if err != nil {
		slog.Error("userDao.FindOneByUserNameForUpdate failed", "openID", openID, "token", token, "ip", ip, "error", err)
		return nil, common.RC_RegisterFailure
	}
	if account != nil {
		slog.Info("RegisterAccount failed, username already exists", "openID", openID, "token", token, "ip", ip)
		return nil, common.RC_UserNameExist
	}
	account, err = srv.doRegisterAccount(openID, token, ip, platform)
	if err != nil {
		return nil, err
	}
	return account, nil
}
func (srv *LoginService) loginOrRegister(openID string, token string, ip string, params url.Values) (*pb.Account, error) {
	now := time.Now().Unix()
	sChannel := params.Get("client")
	channel, _ := strconv.Atoi(sChannel)
	//todo:注册排队限制

	ret := platformVerify(openID, token)
	if ret != nil {
		return nil, ret
	}
	slog.Info("loginOrRegister", "openID", openID, "channel", channel, "ip", ip)

	//平台登陆，用openID当作username，因为user表中OpenID字段不唯一，可为空
	accountDB, err := dao.FindOneByUserNameForUpdate(openID)
	if err != nil {
		slog.Error("userDao.FindOneByUserNameForUpdate failed", "openID", openID, "ip", ip, "error", err)
		return nil, common.RC_LoginFailure
	}
	if accountDB != nil { // 用户已存在
		//todo: 登录限制检查
		slog.Info("LoginOrRegister openID exists", "openID", openID, "channel", channel, "ip", ip)
		accountDB.UserToken = util.CalculateUserToken(accountDB.PassHash + token)
		accountDB.LastIP = ip

		err = dao.UpdateAccountLoginInfo(accountDB.ID, accountDB.LastIP, accountDB.UserToken, true)
		if err != nil {
			slog.Error("userDao.UpdateAccountLoginInfo failed", slog.Group("accountDB", "UserName", accountDB.UserName, "PassHash", accountDB.PassHash, "LastIP", accountDB.LastIP), "error", err)
			return nil, common.RC_LoginFailure
		}
	} else {
		accountDB, err = srv.doRegisterAccount(openID, token, ip, channel)
		if err != nil {
			slog.Error("userDao.AddAccount failed", "openID", openID, "PassHash", accountDB.PassHash, "ip", ip, "channel", channel, "error", err)
			return nil, common.RC_RegisterFailure
		}

	}
	tokenWithTsWithOpenToken := accountDB.UserToken + "|" + strconv.Itoa(int(now)) + "|" + token
	dao.SetAccountToken(accountDB.ID, tokenWithTsWithOpenToken)
	return accountDB, nil
}

// 平台登录校验
func platformVerify(openID string, token string) error {
	return nil
}
