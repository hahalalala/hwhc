package persistence

import (
	"fmt"
	"github.com/go-zhouxun/xutil/xtime"
	"github.com/hwhc/hlc_server/log"
	"github.com/hwhc/hlc_server/mysql"
)

func UpTransferStatus(xmysql *mysql.XMySQL, txDasc string, userId int64, hash string, status, txStatus int64) bool {
	sql := "UPDATE `transactions` SET `tx_status` = ? , `tx_desc` = ?,`update_time` = ?  WHERE `user_id` = ? AND tx_hash = ?  AND `tx_status` = ? "
	result, err := xmysql.Exec(sql, txStatus, txDasc, xtime.TodayDateTimeStr(), userId, hash, status)
	if err != nil {
		fmt.Print("修改提现状态", err, "|userId", userId, "|Id", hash, "|status", status)
		log.Error(fmt.Sprintf("[debug1]UpTransferStatus err : %v,hash:%s,user_id:%d,status:%d,txStatus:%d" , err,hash,userId,status,txStatus))
		return false
	}
	row, err := result.RowsAffected()
	if err != nil || row == 0 { //|| row == 0
		fmt.Print("修改提现状态", err, "|userId", userId, "|Id", hash, "|status", status)
		log.Error(fmt.Sprintf("[debug2]UpTransferStatus SCAN err : %v,hash:%s,user_id:%d,status:%d,txStatus:%d" , err,hash,userId,status,txStatus))
		return false
	}
	return true
}
