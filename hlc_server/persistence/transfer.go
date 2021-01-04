package persistence

import (
	"database/sql"
	"fmt"
	"github.com/go-zhouxun/xutil/xtime"
	"github.com/hwhc/hlc_server/log"
	"github.com/hwhc/hlc_server/mysql"
	"github.com/hwhc/hlc_server/types"
	"strconv"
)

//获取总支出
func GetTotalPay(mysql *mysql.XMySQL, userId, typ int64) float64 {
	sqlstr := "SELECT SUM(amount) FROM `transactions` WHERE `user_id` = ? AND `type` = ? AND `tx_status` = ? and amount < 0"
	row := mysql.QueryRow(sqlstr, userId, typ, 1)
	total := float64(0)
	_ = row.Scan(&total)
	return total
}


//获取总支出
func GetTotalPayAfterRate(mysql *mysql.XMySQL, userId, typ int64) float64 {
	sqlstr := "SELECT SUM(amount * mp_price) FROM `transactions` WHERE `user_id` = ? AND `type` = ? AND `tx_status` = ? and amount < 0"
	row := mysql.QueryRow(sqlstr, userId, typ, 1)
	total := float64(0)
	_ = row.Scan(&total)
	return total
}



//获取交易信息
func GetTransfer(mysql *mysql.XMySQL, txHash string, typ int64) types.Transfer {

	sqlstr := "select `id`,`user_id`, `amount`, `address`, tx_data, `tx_hash` , `tx_status`,`type`,`create_time`,`fee`,`memo`,`tx_desc` , `coin_id`,`is_shop` FROM `transactions` WHERE `tx_hash` = ? AND `type` = ? "
	row := mysql.QueryRow(sqlstr, txHash, typ)

	var transfer types.Transfer
	err := row.Scan(
		&transfer.Id, &transfer.UserId, &transfer.Amount, &transfer.Address, &transfer.Tx_data, &transfer.Tx_hash,
		&transfer.Tx_status, &transfer.Type, &transfer.CreateTime, &transfer.Fee, &transfer.Memo, &transfer.TxDesc,
		&transfer.CoinId, &transfer.IsShop)
	if err != nil && err != sql.ErrNoRows {
		log.Error("GetTransfer err : %v", err)
	}
	return transfer
}

//获取最后免费充值时间
func GetFreeRechargeLastTime(mysql *mysql.XMySQL, userId, coinId int64) string {
	//type：22 就是通过脚本免费给用户充值的
	sqlstr := "select create_time from transactions where user_id = ? and coin_id = ? and type = ? order by id desc limit 1"
	row := mysql.QueryRow(sqlstr, userId, coinId, types.SYSTEM_SCRIPT_FREE_RECHARGE)

	var lastTime string
	err := row.Scan(&lastTime)
	if err != nil && err != sql.ErrNoRows {
		log.Error("GetFreeRechargeLastTime err : %v", err)
	}
	return lastTime
}

func SaveTransfer(mysql *mysql.XMySQL, userId, types int64, coin_id int64, amount float64, address string, data string, tx string, fee float64, memo string, status int64, is_shop int64, mpPrice float64) int64 {
	sql := "INSERT INTO `transactions`(`user_id`, `amount`, `type`, coin_id ,`address`, `tx_data`, `tx_hash`,create_time,`fee`,memo,tx_status,is_shop,mp_price) VALUES(?, ?, ?, ?, ?, ?, ?, ? ,?,?,?,?,?)"

	result, err := mysql.Exec(sql, userId, amount, types, coin_id, address, data, tx, xtime.TodayDateTimeStr(), fee, memo, status, is_shop, mpPrice)
	if err != nil {
		fmt.Print("SaveTransfer insert ", err)
		return -3
	}
	var id int64
	if id, err = result.LastInsertId(); err != nil || id <= 0 {
		fmt.Print("error SaveTransfer , %v, %d", err, id)
		return -4
	}
	return id
}

func TransferbyTxhash(mysql *mysql.XMySQL, userId, cid int64, orderId string) types.Transfer {
	sql := "SELECT `amount`,`fee`,`id`, `tx_hash`, tx_data ,`create_time`, coin_id,`type`, `address`, `tx_status` , `tx_desc`,memo,user_id FROM `transactions` WHERE `user_id` = ? AND `coin_id` = ? AND tx_hash = ? AND tx_status = 0 AND type = 2"
	rows := mysql.QueryRow(sql, userId, cid, orderId)
	var transfer types.Transfer
	err := rows.Scan(&transfer.Amount, &transfer.Fee, &transfer.Id, &transfer.Tx_hash, &transfer.Tx_data, &transfer.CreateTime, &transfer.CoinId, &transfer.Type, &transfer.Address, &transfer.Tx_status, &transfer.TxDesc, &transfer.Memo, &transfer.UserId)

	if err != nil {
		log.Error(fmt.Sprintf("[debug]TransferbyId err ---err : %v userid:%d , cid:%d,orderId:%s", err, userId, cid, orderId))
	}
	return transfer
}

func TransferbyId(mysql *mysql.XMySQL, userId, transferId int64, cid int64) types.Transfer {
	sql := "SELECT `amount`,`fee`,`id`, `tx_hash`, tx_data ,`create_time`, coin_id,`type`, `address`, `tx_status` , `tx_desc`,memo,user_id FROM `transactions` WHERE `user_id` = ? AND `coin_id` = ? and id = ? and tx_status = 0"

	rows := mysql.QueryRow(sql, userId, cid, transferId)
	var transfer types.Transfer
	err := rows.Scan(&transfer.Amount, &transfer.Fee, &transfer.Id, &transfer.Tx_hash, &transfer.Tx_data, &transfer.CreateTime, &transfer.CoinId, &transfer.Type, &transfer.Address, &transfer.Tx_status, &transfer.TxDesc, &transfer.Memo, &transfer.UserId)

	if err != nil {
		log.Error(fmt.Sprintf("[debug]TransferbyId err ---err : %v userid:%d , cid:%d,transferId:%d", err, userId, cid, transferId))
	}
	return transfer
}

func UpdateTransferStatus(mysql *mysql.XMySQL, hash string, user_id int64) bool {

	log.Info(fmt.Sprintf("[debug]UpdateTransferStatus START hash:%s,user_id:%d ", hash, user_id))
	sqlu := "update transactions set tx_status = 1 where user_id = ? and tx_hash = ?  and tx_status = 0 "
	result, err := mysql.Exec(sqlu, user_id, hash)
	if err != nil {
		log.Error(fmt.Sprintf("[debug]UpdateTransferStatus err : %v,hash:%s,user_id:%d", err, hash, user_id))
		return false
	}
	row, err := result.RowsAffected()
	if err != nil || row == 0 { //|| row == 0
		log.Error(fmt.Sprintf("[debug]UpdateTransferStatus RowsAffected err : %v,hash:%s,user_id:%d ,row:%d", err, hash, user_id, row))
		return false
	}
	log.Info(fmt.Sprintf("[debug]UpdateTransferStatus SUCCESS err : %v,hash:%s,user_id:%d ,row:%d", err, hash, user_id, row))
	return true
}

func UpdateTransferHooOrderNo(mysql *mysql.XMySQL, id int64, user_id int64, hooOrderNo string) bool {

	log.Info(fmt.Sprintf("[debug]UpdateTransferHooOrderNo  start ,id :%d ,user_id:%d, hooOrderNo :%s", id, user_id, hooOrderNo))

	sqlu := "update transactions set  hoo_order_no = ?,update_time = ? where user_id = ? and id = ? "
	result, err := mysql.Exec(sqlu, hooOrderNo, xtime.TodayDateTimeStr(), user_id, id)
	if err != nil {
		log.Error(fmt.Sprintf("[debug] [1]UpdateTransferHooOrderNo  err : %v ,id :%d ,user_id:%d, hooOrderNo :%s", err, id, user_id, hooOrderNo))
		return false
	}
	row, err := result.RowsAffected()
	if err != nil || row == 0 { //|| row == 0
		log.Error(fmt.Sprintf("[debug] [2]UpdateTransferHooOrderNo  err : %v ,id :%d ,user_id:%d, hooOrderNo :%s ,row:%d", err, id, user_id, hooOrderNo, row))
		return false
	}

	log.Info(fmt.Sprintf("[debug]UpdateTransferHooOrderNo  SUCCESS id :%d ,user_id:%d, hooOrderNo :%s , row:%d ", id, user_id, hooOrderNo, row))
	return true
}

func TransferList(mysql *mysql.XMySQL, userId, coin_id int64, size, types, lastId int64) []map[string]interface{} {
	sql := "SELECT `amount`,`fee`,`id`, `tx_hash`, tx_data ,`create_time`, coin_id,`type`, `address`, `tx_status` , `tx_desc`,memo FROM `transactions` WHERE `user_id` = ? AND `coin_id` = ? and amount <> 0 and id < ? ORDER BY `id` DESC LIMIT ?"
	params := make([]interface{}, 0)
	params = append(params, userId)
	params = append(params, coin_id)
	if types == 0 { //全部
		sql = "SELECT `amount`,`fee`,`id`, `tx_hash`, tx_data ,`create_time`, coin_id,`type`, `address`, `tx_status` , `tx_desc`,memo FROM `transactions` WHERE `user_id` = ? AND `coin_id` = ? and amount <> 0 and id < ? ORDER BY `id` DESC LIMIT ?"
	} else if types == -1 { //入账
		sql = "SELECT `amount`,`fee`,`id`, `tx_hash`, tx_data ,`create_time`, coin_id,`type`, `address`, `tx_status` , `tx_desc`,memo FROM `transactions` WHERE `user_id` = ? AND `coin_id` = ? and amount > 0 and id < ? ORDER BY `id` DESC LIMIT ?"
	} else if types == -2 { //出账
		sql = "SELECT `amount`,`fee`,`id`, `tx_hash`, tx_data ,`create_time`, coin_id,`type`, `address`, `tx_status` , `tx_desc`,memo FROM `transactions` WHERE `user_id` = ? AND `coin_id` = ? and amount < 0 and id < ? ORDER BY `id` DESC LIMIT ?"
	} else if types > 0 { //根据类型查询
		sql = "SELECT `amount`,`fee`,`id`, `tx_hash`, tx_data ,`create_time`, coin_id,`type`, `address`, `tx_status` , `tx_desc`,memo FROM `transactions` WHERE `user_id` = ? AND `coin_id` = ? and type = ? and id < ? ORDER BY `id` DESC LIMIT ?"
		params = append(params, types)
	}
	params = append(params, lastId)
	params = append(params, size)
	rows, err := mysql.Query(sql, params...)
	if err != nil {
		log.Error("Find transfer list failed, %v", err)
		return nil
	}
	list := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, typ, status, coinid int64
		var amount, fee float64
		var hash, address, txDesc, tx_data, time, memo string
		_ = rows.Scan(&amount, &fee, &id, &hash, &tx_data, &time, &coinid, &typ, &address, &status, &txDesc, &memo)
		//if txDesc != "" {
		//	hash = hash + "驳回 （" + txDesc + "）"
		//}
		m := map[string]interface{}{
			"id":          id,
			"txhash":      hash,
			"create_time": time,
			"coin_id":     coinid,
			"type":        typ,
			"address":     address,
			"amount":      amount,
			"tx_data":     tx_data,
			"fee":         fee,
			"status":      status,
			"memo":        memo,
			"txDesc":      txDesc,
		}
		list = append(list, m)
	}
	return list
}

func TransferCount(mysql *mysql.XMySQL, userId, coin_id int64, types int64) int64 {
	sql := "SELECT count(1) FROM `transactions` WHERE `user_id` = ? AND `coin_id` = ? and amount <> 0 ORDER BY `id` DESC"
	params := make([]interface{}, 0)
	params = append(params, userId)
	params = append(params, coin_id)
	if types == 0 { //全部
		sql = "SELECT count(1) FROM `transactions` WHERE `user_id` = ? AND `coin_id` = ? and amount <> 0 ORDER BY `id` DESC"
	} else if types == -1 { //-1 入账
		sql = "SELECT count(1) FROM `transactions` WHERE `user_id` = ? AND `coin_id` = ? and amount > 0 ORDER BY `id` DESC"
	} else if types == -2 { //出账
		sql = "SELECT count(1) FROM `transactions` WHERE `user_id` = ? AND `coin_id` = ? and amount < 0 ORDER BY `id` DESC"
	} else if types > 0 {
		sql = "SELECT count(1) FROM `transactions` WHERE `user_id` = ? AND `coin_id` = ? and type = ? ORDER BY `id` DESC"
		params = append(params, types)
	}

	row := mysql.QueryRow(sql, params...)

	var nonce int64
	_ = row.Scan(&nonce)
	return nonce
}

func TakeRecordingNum(xmysql *mysql.XMySQL, is_shop int64) int64 {
	sql := "SELECT count(1) FROM transactions  WHERE id > 0  AND tx_status = 0  and is_shop = ? and coin_id != " + strconv.Itoa(IDR)

	rows := xmysql.QueryRow(sql, is_shop)
	var number int64
	err := rows.Scan(&number)
	if err != nil {
		fmt.Println("TakeRecordingNum 查询提现记录总条数错误 err", err)
		log.Error("TakeRecordingNum 查询提现记录总条数错误 err", err)
		return 0
	}
	return number
}

func SelIdTransfer(xmysql *mysql.XMySQL, Id int64) types.Transfer {
	sql := "SELECT id,user_id,amount,address,tx_data,tx_hash,tx_status,type,create_time,fee,coin_id,is_shop FROM transactions WHERE id = ? "
	rows := xmysql.QueryRow(sql, Id)
	var transfer types.Transfer
	err := rows.Scan(&transfer.Id, &transfer.UserId, &transfer.Amount, &transfer.Address, &transfer.Tx_data, &transfer.Tx_hash, &transfer.Tx_status, &transfer.Type, &transfer.CreateTime, &transfer.Fee, &transfer.CoinId, &transfer.IsShop)
	if err != nil {
		log.Error("查询提现记录", err, "user_Id : ", Id)
		fmt.Println("查询提现记录", err, "user_Id :", Id)
	}
	return transfer
}

func TakeRecording(xmysql *mysql.XMySQL, start, end int64, is_shop int64) []types.Transfer {
	sql := "SELECT `id`,`user_id`,`amount`,`address`,`tx_data`,`tx_hash`,`tx_status`,`type`,`create_time`,`fee` ,IFNULL(`update_time`,'0'),coin_id FROM `transactions` WHERE `id` > 0 and tx_status = 0  and is_shop = ?  and coin_id != " + strconv.Itoa(IDR)
	sql = sql + " ORDER BY id DESC LIMIT " + strconv.Itoa(int(start)) + ", " + strconv.Itoa(int(end))
	rows, err := xmysql.Query(sql, is_shop)
	if err != nil {
		log.Error("TakeRecording 查询提现记录错误1", err)
		fmt.Println("TakeRecording 查询提现记录错误1", err)
	}
	lists := make([]types.Transfer, 0)
	for rows.Next() {
		var transfer types.Transfer
		err = rows.Scan(&transfer.Id, &transfer.UserId, &transfer.Amount, &transfer.Address, &transfer.Tx_data, &transfer.Tx_hash, &transfer.Tx_status, &transfer.Type, &transfer.CreateTime, &transfer.Fee, &transfer.UpdateTime, &transfer.CoinId)
		if err != nil {
			log.Error("TakeRecording 查询提现记录Scan错误2", err)
			fmt.Println("TakeRecording 查询提现记录Scan错误2", err)
		}
		lists = append(lists, transfer)
	}
	return lists
}

func GetFuTou(mysql *mysql.XMySQL, beginTime, endTime string) []map[int64]float64 {
	sql1 := " SELECT user_id,sum(amount) from transactions where type=3 and coin_id=1 and amount>0 and tx_status=1 and (create_time BETWEEN ?  and  ?)  GROUP BY user_id"
	rows, err := mysql.Query(sql1, beginTime, endTime)
	if err != nil {
		log.Error("Find transfer list failed, %v", err)
		return nil
	}

	var list []map[int64]float64
	for rows.Next() {
		var userId int64
		var amountTotal float64
		_ = rows.Scan(&userId, &amountTotal)

		m := make(map[int64]float64, 1)
		m[userId] = amountTotal
		list = append(list, m)
	}
	return list
}

func GetReceiveTotal(mysql *mysql.XMySQL, userId int64, startTime, endTime string) float64 {
	sql := "SELECT sum(amount) FROM `transactions` WHERE  user_id = ? and coin_id=1 and amount < 0 and tx_status=1 and (create_time BETWEEN ? and  ?) "
	row := mysql.QueryRow(sql, userId, startTime, endTime)
	var nonce float64
	_ = row.Scan(&nonce)
	return nonce
}
