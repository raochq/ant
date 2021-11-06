package httpService

import (
	"net/url"
	"strconv"
	"time"

	"github.com/raochq/ant/common"
	"github.com/raochq/ant/engine/logger"
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
		logger.Error("GetUniqueAccountID: redis.T64.GetTicket64(%v) failed(%v)", uint(srv.zoneID), err)
		return 0, common.RC_Unknown
	}
	ticket64, err = dao.GenerateTicketPoolID(srv.zoneID)
	if err != nil {
		logger.Error("GetUniqueAccountID: redis.T64.GenerateTicketPoolID(%v) failed(%v)", srv.zoneID, err)
		return 0, common.RC_Unknown
	}
	// todo: id号上限检查
	return ticket64, nil
}

//用户登录
func (srv *LoginService) userLogin(account *pb.Account) error {
	//todo: 登录限制检查
	accountDB, err := dao.FindOneByUserNameForUpdate(account.UserName)
	if err != nil {
		logger.Error("userDao.FindOneByUserNameForUpdate(\"%s\",\"%s\",\"%s\") failed (%s)", account.UserName, account.PassHash, account.LastIP, err.Error())
		return common.RC_AccoutNotExists
	}
	if accountDB == nil {
		return common.RC_AccoutNotExists
	}
	// verify the password
	if account.PassHash != accountDB.PassHash {
		logger.Error("service.Login, user:%s's password not match(%s)", account.UserName, account.PassHash)
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
		logger.Error("RegisterAccount(%v, %v, %v): service.GetUniqueAccountID() failed with error(%v)",
			openID, account.PassHash, ip, ret)
		return nil, ret
	}
	logger.Debug("ticket64: %v", ticket64)
	account.ID = ticket64
	_, err := dao.AddAccount(account)
	if err != nil {
		logger.Error("userDao.AddAccount(%v, %v, %v, %v) failed (%v)", openID, account.PassHash, ip, err)
		return nil, common.RC_RegisterFailure
	}
	return account, nil
}
func (srv *LoginService) registerAccount(openID string, token string, ip string, platform int) (*pb.Account, error) {
	logger.Info("register username:%s, lastIP:%s ", openID, ip)
	account, err := dao.FindOneByUserNameForUpdate(openID)
	if err != nil {
		logger.Error("userDao.FindOneByUserNameForUpdate(\"%s\",\"%s\",\"%s\") failed (%s)", openID, token, ip, err.Error())
		return nil, common.RC_RegisterFailure
	}
	if account != nil {
		logger.Info("RegisterAccount(%v, %v, %v) failed, username already exists", openID, token, ip)
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
	logger.Info("loginOrRegister token:%s, channel:%d, lastIP:%s ", openID, channel, ip)

	//平台登陆，用openID当作username，因为user表中OpenID字段不唯一，可为空
	accountDB, err := dao.FindOneByUserNameForUpdate(openID)
	if err != nil {
		logger.Error("userDao.FindOneByUserNameForUpdate(\"%s\",\"%s\") failed (%s)", openID, ip, err.Error())
		return nil, common.RC_LoginFailure
	}
	if accountDB != nil { // 用户已存在
		//todo: 登录限制检查
		logger.Info("LoginOrRegister(%v, %v, %v) login, openID exists", openID, ip, channel)
		accountDB.UserToken = util.CalculateUserToken(accountDB.PassHash + token)
		accountDB.LastIP = ip

		err = dao.UpdateAccountLoginInfo(accountDB.ID, accountDB.LastIP, accountDB.UserToken, true)
		if err != nil {
			logger.Error("userDao.UpdateAccountLoginInfo(\"%s\",\"%s\",\"%s\") failed (%s)", accountDB.UserName, accountDB.PassHash, accountDB.LastIP, err.Error())
			return nil, common.RC_LoginFailure
		}
	} else {
		accountDB, err = srv.doRegisterAccount(openID, token, ip, channel)
		if err != nil {
			logger.Error("userDao.AddAccount(%v, %v, %v, %v) failed (%s)", openID, accountDB.PassHash, ip, channel, err.Error())
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
