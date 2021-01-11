package persistence

import (
	"fmt"
	"github.com/go-zhouxun/xutil/xtime"
	"github.com/hwhc/hlc_server/log"
	"github.com/hwhc/hlc_server/mysql"
	"github.com/hwhc/hlc_server/types"
)


//获取账户有余额的用户
func GetUserAmountListByLimit(xmysql *mysql.XMySQL, coinId,lastId ,limit int64) ([]types.UserAmount,error) {
	sql := "SELECT `id`,`user_id`,`amount`,`is_shop` FROM `user_amount` WHERE id > ? and `type` = ?  order by id ASC limit ?"
	rows,err := xmysql.Query(sql,  lastId,coinId,limit)
	if err != nil{
		return nil,fmt.Errorf("GetUserAmountListByLimit err:%v,coinId:%d,lastId:%d,limit:%d",err,coinId,lastId,limit)
	}

	var userAmountList = make([]types.UserAmount, 0)
	for rows.Next() {
		var userAmount types.UserAmount
		err = rows.Scan(&userAmount.Id,&userAmount.UserId, &userAmount.Amount,&userAmount.IsShop)
		if err != nil{
			return nil,fmt.Errorf("GetUserAmountListByLimit Scan err : %v,coinId:%d,lastId:%d,limit:%d",err,coinId,lastId,limit)
		}
		userAmountList = append(userAmountList, userAmount)
	}

	return userAmountList,nil
}




//获取账户有余额的用户
func GetHasPriceUserAmountList(xmysql *mysql.XMySQL,  t int64) []types.UserAmount {
	sql := "SELECT `user_id`,`amount` FROM `user_amount` WHERE `type` = ? AND amount > 0"
	rows,err := xmysql.Query(sql,  t)
	if err != nil{
		log.Error("GetHasPriceUserAmountList err : %v,t:%d",err,t)
		return nil
	}

	var userAmountList = make([]types.UserAmount, 0)
	for rows.Next() {
		var userAmount types.UserAmount
		err = rows.Scan(&userAmount.UserId, &userAmount.Amount)
		if err != nil{
			log.Error("GetHasPriceUserAmountList Scan err : %v,t:%d ,userId:%d,amount:%.8f",err,t,userAmount.UserId,userAmount.Amount)
			return nil
		}
		userAmountList = append(userAmountList, userAmount)
	}

	return userAmountList
}


func GetUserAmount(xmysql *mysql.XMySQL, userId, t int64) float64 {
	sql := "SELECT `amount` FROM `user_amount` WHERE `user_id` = ? AND `type` = ?"
	row := xmysql.QueryRow(sql, userId, t)
	amount := float64(0)
	_ = row.Scan(&amount)
	return amount
}

func GetUserFrozenAmount(xmysql *mysql.XMySQL, userId, t int64) (float64, float64) {
	sql := "SELECT `frozen_amount`,tz_amount FROM `user_amount` WHERE `user_id` = ? AND `type` = ?"
	row := xmysql.QueryRow(sql, userId, t)
	amount := float64(0)
	tzAmount := float64(0)
	_ = row.Scan(&amount, &tzAmount)
	return amount, tzAmount
}

func SavePrice(xmysql *mysql.XMySQL, t int64, price float64, time int64, date string) {
	if price <= 0 {
		return
	}
	sql := "INSERT INTO `price`(`type`, `price`, `date_str`, `time`) VALUES(?, ?, ?, ?)"
	//fmt.Print("价格=", price)
	_, err := xmysql.Exec(sql, t, price, date, time)
	if err != nil {
		fmt.Println("SavePrice , failed, %v", err)
	}
}

func AddUserAmount(xmysql *mysql.XMySQL, userId, t int64, amount float64, is_shop int64) bool {
	now := xtime.TodayDateTimeStr()
	sql := "INSERT INTO `user_amount`(`user_id`, `type`, `amount`, `create_time`, `update_time`,is_shop) VALUES(?, ?, ?, ?, ?,?) ON DUPLICATE KEY UPDATE `amount` = `amount` + ?, `update_time` = ?"
	_, err := xmysql.Exec(sql, userId, t, amount, now, now, is_shop, amount, now)
	if err != nil {
		log.Error("add user amount failed, %d, %d, %f, %v", userId, t, amount, err)
		return false
	}
	fmt.Println("增加用户资产成功！～ userId = ", userId, "amount= ", amount)
	return true
}

//添加通政账户基金
func AddUserTZAmount(xmysql *mysql.XMySQL, userId, t int64, TZamount float64) bool {
	now := xtime.TodayDateTimeStr()
	sql := "INSERT INTO `user_amount`(`user_id`, `type`, `tz_amount`, `create_time`, `update_time`) VALUES(?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE `tz_amount` = `tz_amount` + ?, `update_time` = ?"
	_, err := xmysql.Exec(sql, userId, t, TZamount, now, now, TZamount, now)
	if err != nil {
		log.Error("AddUserTZAmount failed, %d, %d, %f, %v", userId, t, TZamount, err)
		return false
	}
	fmt.Println("增加用户通政账户资产成功！～ userId = ", userId, "amount= ", TZamount, "")
	return true
}

//添加冻结资金
func AddUserFrozenAmount(xmysql *mysql.XMySQL, userId, t int64, frozenAmount float64, is_shop int64) bool {
	now := xtime.TodayDateTimeStr()
	sql := "INSERT INTO `user_amount`(`user_id`, `type`, `frozen_amount`, `create_time`, `update_time`,is_shop) VALUES(?, ?, ?, ?, ?,?) ON DUPLICATE KEY UPDATE `frozen_amount` = `frozen_amount` + ?, `update_time` = ?"
	_, err := xmysql.Exec(sql, userId, t, frozenAmount, now, now, is_shop, frozenAmount, now)
	if err != nil {
		log.Error("AddUserTZAmount failed, %d, %d, %f, %v", userId, t, frozenAmount, err)
		return false
	}
	fmt.Println("增加用户通政账户资产成功！～ userId = ", userId, "amount= ", frozenAmount, "")
	return true
}

func ReduceUserAmount(xmysql *mysql.XMySQL, userId, t int64, amount float64) bool {
	now := xtime.TodayDateTimeStr()
	sql := "UPDATE `user_amount` SET `amount` = `amount` - ?, `update_time` = ? WHERE `user_id` = ? AND `type` = ? AND `amount` >= ?"
	result, err := xmysql.Exec(sql, amount, now, userId, t, amount)
	if err != nil {
		fmt.Print("update amount failed ---, %v", err)
		return false
	}
	row, err := result.RowsAffected()
	if err != nil || row == 0 { //|| row == 0
		fmt.Print("reduce user amount failed  ---, %v, %d", err, row)
		return false
	}
	return true
}

func ReduceUserHzAmount(xmysql *mysql.XMySQL, userId, t int64, amount float64) bool {
	now := xtime.TodayDateTimeStr()
	sql := "UPDATE `user_amount` SET `hz_amount` = `hz_amount` - ?, `update_time` = ? WHERE `user_id` = ? AND `type` = ? AND `hz_amount` >= ?"
	result, err := xmysql.Exec(sql, amount, now, userId, t, amount)
	if err != nil {
		fmt.Print("update hz_amount failed ---, %v", err)
		return false
	}
	row, err := result.RowsAffected()
	if err != nil || row == 0 { //|| row == 0
		fmt.Print("reduce user hz_amount failed  ---, %v, %d", err, row)
		return false
	}
	return true
}

//减少冻结资产
func ReduceUserFrozenAmount(xmysql *mysql.XMySQL, userId, t int64, frozen_amount float64) bool {
	now := xtime.TodayDateTimeStr()
	sql := "UPDATE `user_amount` SET `frozen_amount` = `frozen_amount` - ?, `update_time` = ? WHERE `user_id` = ? AND `type` = ? AND `frozen_amount` >= ?"
	//fmt.Println("userId",userId,":t：",t,"-frozen_amount:",frozen_amount)
	result, err := xmysql.Exec(sql, frozen_amount, now, userId, t, frozen_amount)
	if err != nil {
		fmt.Print("ReduceUserFrozenAmount failed ---, %v", err, " userId=", userId, " type=", t, " frozen_amount=", frozen_amount)
		log.Error("ReduceUserFrozenAmount failed ---, %v", err, " userId=", userId, " type=", t, " frozen_amount=", frozen_amount)
		return false
	}
	row, err := result.RowsAffected()
	if err != nil || row == 0 { //|| row == 0
		fmt.Print("ReduceUserFrozenAmount failed  ---, %v, %d", err, row, " userId=", userId, " type=", t, " frozen_amount=", frozen_amount)
		log.Error("ReduceUserFrozenAmount failed  ---, %v, %d", err, row, " userId=", userId, " type=", t, " frozen_amount=", frozen_amount)
		return false
	}
	return true
}

func GetNonce(xmysql *mysql.XMySQL) uint64 {
	sql := "SELECT `value` FROM `conf` WHERE `id` = 2"
	nonce := uint64(0)
	row := xmysql.QueryRow(sql)
	_ = row.Scan(&nonce)
	return nonce
}

func AddNonce(xmysql *mysql.XMySQL) {
	sql := "UPDATE `conf` SET `value` = `value` + 1 WHERE `id` = 2"
	_, _ = xmysql.Exec(sql)
}

func GetCoinMarketPath(xmysql *mysql.XMySQL, CoinType int64) string {
	sql := "SELECT `market_path` from coin where id = ?"
	row := xmysql.QueryRow(sql, CoinType)
	if row == nil {
		return ""
	}
	url := ""
	_ = row.Scan(&url)
	return url
}

func GetCoinInfo(xmysql *mysql.XMySQL, CoinType int64) types.CoinPirce {
	var ss types.CoinPirce
	sql := "SELECT `coinname`,`url`,`sortname` from coin where id = ?"
	row := xmysql.QueryRow(sql, CoinType)
	//if row==nil{
	//	return ss
	//}
	_ = row.Scan(&ss.CoinName, &ss.Icon, &ss.SortName)
	return ss
}

func FindUserRecharge(xmysql *mysql.XMySQL, userid int64, token string) []types.User_recharge {

	sql := "select time, amount from `recharge_record` where user_id = ? AND "

	var findsql string
	if token == "HC" {
		findsql = sql + "coinname is null"
	} else {
		findsql = fmt.Sprintf("%s%s%s%s%s", sql, "coinname = ", `"`, token, `"`)
	}

	rows, err := xmysql.Query(findsql, userid)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	var rechargelist = make([]types.User_recharge, 0)
	for rows.Next() {
		var recharge types.User_recharge
		rows.Scan(&recharge.Time, &recharge.Amount)
		recharge.Tokenname = token

		rechargelist = append(rechargelist, recharge)
	}
	fmt.Println(rechargelist)
	return rechargelist
}

func WalletRechargeRecord(xmysql *mysql.XMySQL, userid int64, coinid int64) ([]types.WalletRechargeRecord, int64) {
	sql := "select coinid , amount , status , time , tx_hash, `desc`,`fee`  from  `recharge_record` where user_id = ?  and coinid = ? order by id desc"
	rows, err := xmysql.Query(sql, userid, coinid)
	if err != nil {
		log.Error("WalletRechargeRecord select err ", err)
		return []types.WalletRechargeRecord{}, -1
	}
	var lists = make([]types.WalletRechargeRecord, 0)
	for rows.Next() {
		var w types.WalletRechargeRecord
		err := rows.Scan(&w.CoinId, &w.Amount, &w.Status, &w.Time, &w.Txhash, &w.Desc, &w.Fee)
		if err != nil {
			log.Error("WalletRechargeRecord select err ", err)
			return []types.WalletRechargeRecord{}, -2
		}
		lists = append(lists, w)
	}
	return lists, 1
}

func FindRechargeByOrderId(xmysql *mysql.XMySQL, txHash string) types.WalletRechargeRecord_call {
	sql := "select coinid , amount , status , time , tx_hash,user_id from `recharge_record` where tx_hash = ? "
	row := xmysql.QueryRow(sql, txHash)

	var w types.WalletRechargeRecord_call
	errs := row.Scan(&w.CoinId, &w.Amount, &w.Status, &w.Time, &w.Txhash, &w.UserId)
	if errs != nil {
		log.Error("findRechargeByOrderId select err 111", errs)
		return w
	}

	return w
}

func Insert24TimeRecordCoin(xmysql *mysql.XMySQL, price float64, mtype int64) {

	sql := "INSERT INTO date_price (`date_str`,`max_price`,`time`,`confirmed`,`token_type`) VALUES (?,?,?,?,?)"
	_, err := xmysql.Exec(sql, xtime.TodayDateStr(), price, xtime.Now(), 1, mtype)
	if err != nil {
		log.Error("add date_price failed", err)
	}
}
