package persistence

import (
	"fmt"
	"github.com/hwhc/hlc_server/mysql"
	"github.com/hwhc/hlc_server/types"
	"github.com/hwhc/hlc_server/util"
)

//添加增值记录
func AddAmountIncrRecord(xmysql  *mysql.XMySQL,userId,coinId,isShop int64,amount,incrValue,nowAmountPrice,beforeAmountPrice float64,incrNowDate ,incrBeforeDate string) error {

	sqlstr := "INSERT INTO amount_incr_record (user_id,coin_id,amount,incr_value,now_amount_price,before_amount_price,is_shop,create_date,incr_now_date,incr_before_date) VALUES (?,?,?,?,?,?,?,?,?,?)"
	result,err := xmysql.Exec(sqlstr,userId,coinId,amount,incrValue,nowAmountPrice,beforeAmountPrice,isShop,util.Datestr(),incrNowDate,incrBeforeDate)
	if err != nil{
		return fmt.Errorf("AddAmountIncrRecord Exec sql err:%v , userId:%d ,coinId:%d ,amount:%.8f,incrValue:%.8f,nowAmountPrice:%.8f,beforeAmountPrice:%.8f",
			err,userId,coinId,amount,incrValue,nowAmountPrice,beforeAmountPrice)
	}
	if id, err := result.LastInsertId(); err != nil || id <= 0 {
		return fmt.Errorf("AddAmountIncrRecord LastInsertId  err:%v , userId:%d ,coinId:%d ,amount:%.8f,incrValue:%.8f,nowAmountPrice:%.8f,beforeAmountPrice:%.8f,incr_now_date:%s,incr_before_date:%s,",
			err,userId,coinId,amount,incrValue,nowAmountPrice,beforeAmountPrice,incrNowDate,incrBeforeDate)
	}

	return nil
}


func GetAmountIncrRecordId(xmysql  *mysql.XMySQL,userId,coinId int64,dateStr string)int64 {
	sqlstr := "SELECT id FROM amount_incr_record WHERE user_id = ? AND coin_id = ? AND incr_now_date = ?"
	row :=xmysql.QueryRow(sqlstr,userId,coinId,dateStr)
	var id int64
	_ = row.Scan(&id)
	return id
}

//获取总的累计增值
func GetAmountIncrTotal(xmysql  *mysql.XMySQL,userId,coinId int64) float64 {
	sqlstr := "SELECT sum(incr_value) FROM amount_incr_record WHERE user_id = ? AND coin_id = ? "
	row :=xmysql.QueryRow(sqlstr,userId , coinId)
	var total float64
	_ = row.Scan(&total)
	return total
}

//根据时间获取增值
func GetAmountIncrByDate(xmysql  *mysql.XMySQL,userId ,coinId int64,dateStr string) float64 {
	sqlstr := "SELECT incr_value FROM amount_incr_record WHERE user_id = ? AND coin_id = ? AND create_date = ? "
	row :=xmysql.QueryRow(sqlstr,userId,coinId,dateStr)
	var incrValue float64
	_ = row.Scan(&incrValue)
	return incrValue
}


//根据时间获取增值
func GetLastAmountIncr(xmysql  *mysql.XMySQL,userId ,coinId int64) map[string]interface{} {
	sqlstr := "SELECT incr_value,now_amount_price FROM amount_incr_record WHERE user_id = ? AND coin_id = ? ORDER BY id DESC LIMIT 1 "
	row :=xmysql.QueryRow(sqlstr,userId,coinId)
	var incrValue float64
	var nowAmountPrice float64
	r := make(map[string]interface{})
	_ = row.Scan(&incrValue,&nowAmountPrice)

	r["incrValue"] = incrValue
	r["nowAmountPrice"] = nowAmountPrice

	return r
}


//获取记录数量
func GetAmountIncrRecordCount(xmysql  *mysql.XMySQL,userId ,coinId int64,startDateStr,endDateStr string)(total int64,err error)  {
	sqlstr := "SELECT count(id) FROM  amount_incr_record WHERE user_id = ? AND coin_id = ?"

	args := make([]interface{},0)
	args = append(args,userId,coinId)
	if len(startDateStr) > 0{
		sqlstr += " and (create_date >= ?)"
		args = append(args,startDateStr)
	}

	if len(endDateStr) > 0{
		sqlstr += " and (create_date <= ?)"
		args = append(args,endDateStr)
	}


	row := xmysql.QueryRow(sqlstr,args...)
	err  = row.Scan(&total)
	return
}

//获取记录列表
func GetAmountIncrRecordList(xmysql  *mysql.XMySQL,userId ,coinId int64,startDateStr,endDateStr string,lastId,limit int64)(results []types.AmountIncrRecord,err error)  {

	sqlstr := "SELECT id,user_id,coin_id,amount,incr_value,now_amount_price,before_amount_price,is_shop,create_time,update_time FROM  amount_incr_record WHERE user_id = ? AND coin_id = ?"

	//codition
	args := make([]interface{},0)
	args = append(args,userId,coinId)
	if len(startDateStr) > 0{
		sqlstr += " AND (create_date >= ?)"
		args = append(args,startDateStr)
	}
	if len(endDateStr) > 0{
		sqlstr += " AND (create_date <= ?)"
		args = append(args,endDateStr)
	}

	//page
	if lastId > 0{
		sqlstr += " AND id < ? "
		args = append(args,lastId)
	}
	sqlstr += " ORDER BY id DESC LIMIT ?"
	args = append(args,limit)


	rows,err := xmysql.Query(sqlstr,args...)
	if err != nil{
		return results,fmt.Errorf("persistence GetAmountIncrRecordList Query err : %v ,userId:%d ,coinId:%d ,startDate:%s,endDate:%s",err,userId,coinId,startDateStr,endDateStr)
	}

	for rows.Next(){
		var record types.AmountIncrRecord
		err  = rows.Scan(&record.Id,&record.UserId,&record.CoinId,&record.Amount,&record.IncrValue,&record.NowAmountPrice,&record.BeforeAmountPrice,&record.IsShop,&record.CreateTime,&record.UpdateTime)
		if err != nil{
			return nil,fmt.Errorf("persistence GetAmountIncrRecordList Scan err : %v ,userId:%d ,coinId:%d ,startDate:%s,endDate:%s",err,userId,coinId,startDateStr,endDateStr)
		}

		results = append(results,record)
	}

	return
}

