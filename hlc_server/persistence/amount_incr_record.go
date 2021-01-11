package persistence

import (
	"fmt"
	"github.com/hwhc/hlc_server/mysql"
)

//添加增值记录
func AddAmountIncrRecord(xmysql  *mysql.XMySQL,userId,coinId,isShop int64,amount,incrValue,nowAmountPrice,beforeAmountPrice float64) error {

	sqlstr := "INSERT INTO amount_incr_record (user_id,coin_id,amount,incr_value,now_amount_price,before_amount_price,is_shop) VALUES (?,?,?,?,?,?,?)"
	result,err := xmysql.Exec(sqlstr,userId,coinId,amount,incrValue,nowAmountPrice,beforeAmountPrice)
	if err != nil{
		return fmt.Errorf("AddAmountIncrRecord Exec sql err:%v , userId:%d ,coinId:%d ,amount:%.8f,incrValue:%.8f,nowAmountPrice:%.8f,beforeAmountPrice:%.8f",
			err,userId,coinId,amount,incrValue,nowAmountPrice,beforeAmountPrice)
	}
	if id, err := result.LastInsertId(); err != nil || id <= 0 {
		return fmt.Errorf("AddAmountIncrRecord LastInsertId  err:%v , userId:%d ,coinId:%d ,amount:%.8f,incrValue:%.8f,nowAmountPrice:%.8f,beforeAmountPrice:%.8f",
			err,userId,coinId,amount,incrValue,nowAmountPrice,beforeAmountPrice)
	}

	return nil
}