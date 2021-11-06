package dao

import "strconv"

const (
	RedisServerTypeZone  = 1 // 本区
	RedisServerTypeCross = 2 // 跨区
)
const (
	TypeUser    = "user"
	TypeAccount = "account"
)

// account
const (
	ZoneAccountCnt  = 1000000000
	RobotAccountMax = 200000000
)

// For TypeAccount, TypeGuild
const (
	SpecTicket          = "tid"
	SpecMsgQ            = "msgq"
	SpecMsgQCache       = "msgqcache"
	SpecActionMsgQ      = "amsgq"
	SpecActionMsgQCache = "amsgqcache"
	SepcCTime           = "ctime"
)

func Gen(kt, ks string, id uint64) string {
	return kt + ":" + strconv.FormatUint(id, 10) + ":" + ks
}
