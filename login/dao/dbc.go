package dao

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	_ "github.com/go-sql-driver/mysql"
	"github.com/raochq/ant/protocol/pb"
)

const (
	InsertAccountSQL                 = "INSERT IGNORE INTO account(AccountID, UserName, PassHash, UserToken, LastIP, Platform, CTime, MTime, ChatNeteaseToken, IsLoggedIn) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
	GetAccountByUserNameSQLForUpdate = "SELECT AccountID, UserName, PassHash, UserToken, LastIP, IsForbidden, Forbiddenreason, FTime, CTime, MTime, ChatNeteaseToken FROM account WHERE UserName = ? "
	UpdateAccountLoginStatusSQL      = "UPDATE account SET UserToken = ?, IsLoggedIn = ?, LastIP = ? WHERE AccountID = ?"
)

var (
	gDBDao *sql.DB
)

// 初始化
func InitMysql(dbAddr string, maxConn int32) error {
	var err error
	if gDBDao, err = sql.Open("mysql", dbAddr); err != nil {
		slog.Error("sql.Open failed", "addr", dbAddr, "error", err)
		return err
	}
	gDBDao.SetMaxIdleConns(0)
	gDBDao.SetMaxOpenConns(int(maxConn))
	return nil
}

// 获取单个用户
func FindOneByUserNameForUpdate(userName string) (*pb.Account, error) {
	account := &pb.Account{}
	row := gDBDao.QueryRow(GetAccountByUserNameSQLForUpdate, userName)
	if err := row.Scan(&account.ID, &account.UserName, &account.PassHash, &account.UserToken, &account.LastIP,
		&account.IsForbidden, &account.ForbidReason, &account.FTime, &account.CTime, &account.MTime, &account.ChatNeteaseToken); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		} else {
			slog.Error("row.Scan() failed", "error", err)
			return nil, err
		}

	}

	return account, nil
}

// 更新账户登录信息
func UpdateAccountLoginInfo(accountID int64, lastIP string, userToken string, isLoggedIn bool) error {
	_, err := gDBDao.Exec(UpdateAccountLoginStatusSQL, userToken, isLoggedIn, lastIP, accountID)
	if err != nil {
		slog.Error("db.Exec() failed", "error", err)
		return err
	}

	return nil
}

// 添加用户
func AddAccount(account *pb.Account) (int64, error) {
	res, err := gDBDao.Exec(InsertAccountSQL, account.ID, account.UserName, account.PassHash, account.UserToken, account.LastIP,
		account.Platform, account.CTime, account.MTime, account.ChatNeteaseToken, account.IsLoggedIn)
	if err != nil {
		slog.Error("db.Exec() failed", "error", err.Error())
		return 0, err
	}
	row, err := res.RowsAffected()
	if err != nil {
		slog.Error("res.RowsAffected() failed", "error", err.Error())
		return 0, err
	} else if row != 1 {
		slog.Error("res.RowsAffected() row != 1", "rows", row)
		return 0, errors.New(fmt.Sprintf("AddAccount(%v, %v) hasn't insert succeessfully", account, account.ID))
	}

	return row, nil
}
