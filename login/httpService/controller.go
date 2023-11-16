package httpService

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/raochq/ant/common"
	"github.com/raochq/ant/protocol/pb"
)

// POST 返回统一json
func retPWriter(r *http.Request, wr http.ResponseWriter, params *string, start time.Time, result map[string]interface{}) {
	byteJson, err := json.Marshal(result)
	if err != nil {
		slog.Error("json.Marshal failed", "data", result, "error", err)
	}
	timeInServerEnd := time.Now()
	if _, err = wr.Write(byteJson); err != nil {
		slog.Error("wr.Write failed", "data", string(byteJson), "error", err)
	}

	// Log
	if ret, ok := result["ret"]; ok {
		slog.Info("retPWriter ok", "addr", r.RemoteAddr, "url", r.URL.String(),
			"param", *params, "time", time.Now().Sub(start).Seconds(), "ptime", timeInServerEnd.Sub(start).Seconds(), "ret", ret)
	}
}
func getIP(r *http.Request) string {
	slog.Debug("getIP", "Request", r)
	remoteAddr := r.Header.Get("X-Forwarded-For")
	remoteIP := strings.Split(remoteAddr, ",")

	slog.Debug("remoteIP", "ip", remoteIP)
	ip := ""
	ip = remoteIP[0]
	if ip != "" {
		return ip
	}
	ips := r.RemoteAddr
	slog.Debug("getIP RemoteAddr", "r.RemoteAddr", ips)
	rip := strings.Split(ips, ":")
	if len(rip) != 0 {
		ip = rip[0]
		slog.Debug("getIP", "ip", ip)
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
		slog.Error("Method Not Allowed", "Method", r.Method)
		http.Error(wr, "Method Not Allowed", 405)
		return
	}
	res := map[string]interface{}{}
	pStr := ""
	defer retPWriter(r, wr, &pStr, time.Now(), res)

	//todo: 服务状态检查。

	body, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Warn("io.ReadAll failed", "error", err.Error())
		res["ret"] = common.RC_ParameterInvalid
		return
	}
	pStr = string(body)
	params, err := url.ParseQuery(pStr)
	if err != nil {
		slog.Error("url.ParseQuery failed", "query", string(body), "error", err.Error())
		res["ret"] = common.RC_ParameterInvalid
		return
	}
	account := &pb.Account{}
	platType := params.Get("client")
	bpt, err := strconv.Atoi(platType)
	if err != nil {
		slog.Info("strconv.Atoi failed", "str", platType, "error", err.Error())
		res["ret"] = common.RC_ParameterInvalid
		return
	}
	account.Platform = uint32(bpt)
	account.UserName = params.Get("user_name")
	if account.UserName == "" {
		slog.Info("user name is empty")
		res["ret"] = common.RC_ParameterInvalid
		return
	}
	account.PassHash = params.Get("pass_hash")
	if account.PassHash == "" {
		slog.Info("password hash is empty")
		res["ret"] = common.RC_ParameterInvalid
		return
	}

	account.LastIP = getIP(r)

	slog.Debug("UserLogin account %s", account.UserName)
	ret := srv.userLogin(account)
	if ret != nil {
		slog.Info("UserLogin account get user from db fail", "ret", ret)
		res["ret"] = ret
		return
	}

	slog.Debug("UserLogin account get account data from db success", "UserName", account.UserName)
	res["data"] = map[string]string{
		"account_id": fmt.Sprint(account.ID),
		"user_token": account.UserToken,
	}
	res["ret"] = 0
}
func (srv *LoginService) UserRegister(wr http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		slog.Error("Method Not Allowed", "method", r.Method)
		http.Error(wr, "Method Not Allowed", 405)
		return
	}
	res := map[string]interface{}{}
	pStr := ""
	defer retPWriter(r, wr, &pStr, time.Now(), res)

	//todo: 服务状态检查。

	body, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Warn("io.ReadAll failed", "error", err.Error())
		res["ret"] = common.RC_ParameterInvalid
		return
	}
	pStr = string(body)
	params, err := url.ParseQuery(pStr)
	if err != nil {
		slog.Error("url.ParseQuery failed", "query", string(body), "error", err.Error())
		res["ret"] = common.RC_ParameterInvalid
		return
	}
	userName := params.Get("user_name")
	if userName == "" {
		slog.Info("user name is empty")
		res["ret"] = common.RC_ParameterInvalid
		return
	}
	passWord := params.Get("user_pass")
	if passWord == "" {
		slog.Info("password is empty")
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

// 用户登录或注册并登录
func (srv *LoginService) UserLoginOrRegister(wr http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		slog.Error("Method Not Allowed", "method", r.Method)
		http.Error(wr, "Method Not Allowed", 405)
		return
	}
	res := map[string]interface{}{}
	pStr := ""
	defer retPWriter(r, wr, &pStr, time.Now(), res)

	//todo: 服务状态检查限制。

	body, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Warn("io.ReadAll failed", "error", err.Error())
		res["ret"] = common.RC_ParameterInvalid
		return
	}
	pStr = string(body)
	params, err := url.ParseQuery(pStr)
	if err != nil {
		slog.Error("url.ParseQuery failed", "query", string(body), "error", err.Error())
		res["ret"] = common.RC_ParameterInvalid
		return
	}
	userName := params.Get("user_name")
	if userName == "" {
		slog.Info("user name is empty")
		res["ret"] = common.RC_ParameterInvalid
		return
	}
	passWord := params.Get("user_pass")
	if passWord == "" {
		slog.Info("password is empty")
		res["ret"] = common.RC_ParameterInvalid
		return
	}

	openID := params.Get("openID")
	if openID == "" {
		slog.Error("openID is empty")
		res["ret"] = common.RC_ParameterInvalid
		return
	}

	token := params.Get("token")
	if token == "" {
		slog.Error("token is empty")
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
