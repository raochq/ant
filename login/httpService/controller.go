package httpService

import (
	"encoding/json"
	"fmt"
	"github.com/raochq/ant/common"
	"github.com/raochq/ant/engine/logger"
	"github.com/raochq/ant/protocol/pb"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// POST 返回统一json
func retPWriter(r *http.Request, wr http.ResponseWriter, params *string, start time.Time, result map[string]interface{}) {
	byteJson, err := json.Marshal(result)
	if err != nil {
		logger.Error("json.Marshal(\"%v\") failed (%s)", result, err.Error())
	}
	timeInServerEnd := time.Now()
	if _, err = wr.Write(byteJson); err != nil {
		logger.Error("wr.Write(\"%s\") failed (%s)", string(byteJson), err.Error())
	}

	// Log
	if ret, ok := result["ret"]; ok {
		logger.Info("[%s]get_url:%s(param:%s,time:%f,ptime:%f,ret:%d)", r.RemoteAddr, r.URL.String(), *params, time.Now().Sub(start).Seconds(), timeInServerEnd.Sub(start).Seconds(), ret)
	}
}
func getIP(r *http.Request) string {
	logger.Debug("r %v", r)
	remoteAddr := r.Header.Get("X-Forwarded-For")
	remoteIP := strings.Split(remoteAddr, ",")

	logger.Debug("remoteIP %v", remoteIP)
	ip := ""
	logger.Debug("len(remoteIP) : %v", len(remoteIP))
	ip = remoteIP[0]
	if ip != "" {
		return ip
	}
	ips := r.RemoteAddr
	logger.Debug("r.RemoteAddr %v", ips)
	rip := strings.Split(ips, ":")
	if len(rip) != 0 {
		ip = rip[0]
		logger.Debug("ip %v", ip)
		return ip
	}
	return ""
}

func (srv *LoginService) NullRequest(http.ResponseWriter, *http.Request) {
	return
}

// 用户登录http， 必须是post请求
func (srv *LoginService) UserLogin(wr http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		logger.Error("Method Not Allowed %v", r.Method)
		http.Error(wr, "Method Not Allowed", 405)
		return
	}
	res := map[string]interface{}{}
	pStr := ""
	defer retPWriter(r, wr, &pStr, time.Now(), res)

	//todo: 服务状态检查。

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logger.Warn("ioutil.ReadAll() failed (%s)", err.Error())
		res["ret"] = common.RC_ParameterInvalid
		return
	}
	pStr = string(body)
	params, err := url.ParseQuery(pStr)
	if err != nil {
		logger.Error("url.ParseQuery(\"%s\") failed (%s)", string(body), err.Error())
		res["ret"] = common.RC_ParameterInvalid
		return
	}
	account := &pb.Account{}
	platType := params.Get("client")
	bpt, err := strconv.Atoi(platType)
	if err != nil {
		logger.Info("strconv.Atoi(\"%s\") failed (%s)", platType, err.Error())
		res["ret"] = common.RC_ParameterInvalid
		return
	}
	account.Platform = uint32(bpt)
	account.UserName = params.Get("user_name")
	if account.UserName == "" {
		logger.Info("user name is empty")
		res["ret"] = common.RC_ParameterInvalid
		return
	}
	account.PassHash = params.Get("pass_hash")
	if account.PassHash == "" {
		logger.Info("password hash is empty")
		res["ret"] = common.RC_ParameterInvalid
		return
	}

	account.LastIP = getIP(r)

	logger.Debug("UserLogin account %s", account.UserName)
	ret := srv.userLogin(account)
	if ret != nil {
		logger.Info("UserLogin account get user from db fail %d", ret)
		res["ret"] = ret
		return
	}

	logger.Debug("UserLogin account get account %s data from db success", account.UserName)
	res["data"] = map[string]string{
		"account_id": fmt.Sprint(account.ID),
		"user_token": account.UserToken,
	}
	res["ret"] = 0
}
func (srv *LoginService) UserRegister(wr http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		logger.Error("Method Not Allowed %v", r.Method)
		http.Error(wr, "Method Not Allowed", 405)
		return
	}
	res := map[string]interface{}{}
	pStr := ""
	defer retPWriter(r, wr, &pStr, time.Now(), res)

	//todo: 服务状态检查。

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logger.Warn("ioutil.ReadAll() failed (%s)", err.Error())
		res["ret"] = common.RC_ParameterInvalid
		return
	}
	pStr = string(body)
	params, err := url.ParseQuery(pStr)
	if err != nil {
		logger.Error("url.ParseQuery(\"%s\") failed (%s)", string(body), err.Error())
		res["ret"] = common.RC_ParameterInvalid
		return
	}
	userName := params.Get("user_name")
	if userName == "" {
		logger.Info("user name is empty")
		res["ret"] = common.RC_ParameterInvalid
		return
	}
	passWord := params.Get("user_pass")
	if passWord == "" {
		logger.Info("password is empty")
		res["ret"] = common.RC_ParameterInvalid
		return
	}
	sChannel := params.Get("client")
	channel, _ := strconv.Atoi(sChannel)
	currentIP := getIP(r)
	_, ret := srv.registerAccount(userName, passWord, currentIP, channel)
	if ret != nil {
		res["ret"] = ret
		return
	}

	res["ret"] = 0
}

//用户登录或注册并登录
func (srv *LoginService) UserLoginOrRegister(wr http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		logger.Error("Method Not Allowed %v", r.Method)
		http.Error(wr, "Method Not Allowed", 405)
		return
	}
	res := map[string]interface{}{}
	pStr := ""
	defer retPWriter(r, wr, &pStr, time.Now(), res)

	//todo: 服务状态检查限制。

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logger.Warn("ioutil.ReadAll() failed (%s)", err.Error())
		res["ret"] = common.RC_ParameterInvalid
		return
	}
	pStr = string(body)
	params, err := url.ParseQuery(pStr)
	if err != nil {
		logger.Error("url.ParseQuery(\"%s\") failed (%s)", string(body), err.Error())
		res["ret"] = common.RC_ParameterInvalid
		return
	}
	userName := params.Get("user_name")
	if userName == "" {
		logger.Info("user name is empty")
		res["ret"] = common.RC_ParameterInvalid
		return
	}
	passWord := params.Get("user_pass")
	if passWord == "" {
		logger.Info("password is empty")
		res["ret"] = common.RC_ParameterInvalid
		return
	}

	openID := params.Get("openID")
	if openID == "" {
		logger.Error("openID is empty")
		res["ret"] = common.RC_ParameterInvalid
		return
	}

	token := params.Get("token")
	if token == "" {
		logger.Error("token is empty")
		res["ret"] = common.RC_ParameterInvalid
		return
	}
	currentIP := getIP(r)
	account, ret := srv.loginOrRegister(openID, token, currentIP, params)
	if ret != nil {
		res["ret"] = ret
		return
	}
	res["data"] = map[string]string{
		"account_id": fmt.Sprint(account.ID),
		"user_token": account.UserToken,
		//"portal":     portal,
	}
	res["ret"] = 0
}
