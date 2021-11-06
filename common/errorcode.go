package common

import "fmt"

type ErrorCode int32

func (e ErrorCode) Error() string {
	return fmt.Sprintf("Error(%v)", e)
}

const (
	RC_Unknown           ErrorCode = -1    // 未知错误
	RC_ParameterInvalid  ErrorCode = -2    // 参数错误
	RC_UserNameExist     ErrorCode = -1000 // 用户名已存在
	RC_UserNotExist      ErrorCode = -1001 // 用户不存在
	RC_PasswordInvalid   ErrorCode = -1002 // 密码不正确
	RC_TokenInvalid      ErrorCode = -1003 // Token无效(验证未通过)
	RC_LogoutFailure     ErrorCode = -1004 // 用户退出失败
	RC_RegisterFailure   ErrorCode = -1005 // 注册失败
	RC_LoginFailure      ErrorCode = -1014 // 登陆失败
	RC_AccoutNotExists   ErrorCode = -1022 // 账户不存在
	RC_RPCAccoutNotLogin ErrorCode = -1023 // 账户未登录
)
