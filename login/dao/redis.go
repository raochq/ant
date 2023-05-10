package dao

import (
	"time"

	"github.com/gomodule/redigo/redis"

	"github.com/raochq/ant/config"
	"github.com/raochq/ant/util/logger"
)

var (
	pool      *redis.Pool
	crossPool *redis.Pool
)

func GetTicket64(zoneID uint32) (int64, error) {
	con := GetPool(RedisServerTypeZone).Get()
	defer con.Close()
	ticketKey := Gen(TypeAccount, SpecTicket, uint64(zoneID))
	ticket, err := redis.Int64(con.Do("GET", ticketKey))
	if err != nil {
		logger.Error("can't get ticket id from redis for key %s with error %v", ticketKey, err)
		return 0, err
	}
	logger.Info("GetTicket64, %v", ticket)

	return ticket, nil
}
func GenerateTicketPoolID(zoneID uint32) (int64, error) {
	con := GetPool(RedisServerTypeZone).Get()
	defer con.Close()
	ticketKey := Gen(TypeAccount, SpecTicket, uint64(zoneID))
	ticket, err := redis.Int64(con.Do("INCR", ticketKey))
	if err != nil {
		logger.Error("can't incr ticket id from redis for key %s with error %v", ticketKey, err)
		return 0, err
	}

	logger.Info("GenerateTicketPoolID, %v", ticket)
	return ticket, nil
}

func InitTicket64(zoneID uint32) error {
	// get the ticket from redis
	con := GetPool(RedisServerTypeZone).Get()
	defer con.Close()
	ticketKey := Gen(TypeAccount, SpecTicket, uint64(zoneID))
	ticket0 := int64(zoneID) * ZoneAccountCnt
	_, err := con.Do("SETNX", ticketKey, ticket0)
	if err != nil {
		return err
	}
	return nil
}

func SetAccountToken(accountID int64, token string) {
	conn := GetPool(RedisServerTypeZone).Get()
	defer conn.Close()
	rKey := Gen(TypeUser, "token", uint64(accountID))
	_, err := conn.Do("Set", rKey, token)
	if err != nil {
		logger.Warn("set account token error %s", err.Error())
	}
}

func InitRedis(cfg config.Redis) {
	pool = newRedisPool(cfg.Idle, cfg.Timeout, cfg.Addr, cfg.Auth, cfg.Index)
	crossPool = newRedisPool(cfg.Idle, cfg.Timeout, cfg.CrosseAddr, cfg.CrosseAuth, cfg.CrosseIndex)
}

func newRedisPool(maxIdle, timeout int32, addr, auth string, index int32) *redis.Pool {
	p := &redis.Pool{
		MaxIdle:     int(maxIdle),
		IdleTimeout: time.Duration(timeout) * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", addr)
			if err != nil {
				logger.Error("ZoneRedis dial fail %s", err.Error())
				return nil, err
			}
			if auth != "" {
				if _, err := c.Do("AUTH", auth); err != nil {
					logger.Error("c.Do('AUTH', %v) failed(%v)", auth, err)
					c.Close()
					return nil, err
				}
			}
			if index > 0 && index < 16 {
				if _, err = c.Do("SELECT", index); err != nil {
					logger.Error("c.Do('SELECT', %v) failed(%v)", index, err)
					c.Close()
					return nil, err
				}
			}
			logger.Info("ZoneRedis dial success %v %v %v", addr, auth, index)
			return c, nil
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) < time.Second {
				return nil
			}
			_, err := c.Do("PING")
			return err
		},
	}
	return p
}

func GetPool(redisType int) *redis.Pool {
	switch redisType {
	case RedisServerTypeCross:
		return crossPool
	default:
		return pool
	}
}
