package persistence

import (
	"fmt"
	"github.com/hwhc/hlc_server/types"
	"github.com/hwhc/hlc_server/hoo"
	"github.com/hwhc/hlc_server/log"
	"github.com/hwhc/hlc_server/mysql"
	"strconv"
)

func CreateAddress(xmysql *mysql.XMySQL, coinname string, num string) bool {
	var i int64
	i = 0
	num_int, _ := strconv.ParseInt(num, 10, 64)
	for num_int > i {
		addresses := hoo.CreateHcAddress(coinname, "1")
		for address := range addresses {
			println("地址创建成功..", addresses[address])
			sql := "INSERT INTO `address_pool`(`address`, `status`, `coinname`, `user_id`) VALUES(?, ?, ?, ?)"
			result, err := xmysql.Exec(sql, addresses[address], 0, coinname, 0)
			if err != nil {
				log.Error("create address, roll back, %v", err)
				return false
			}
			id, err := result.LastInsertId()
			if err != nil || id <= 0 {
				log.Error("create address, roll back, %v, %d", err, id)
				return false
			}

		}
		i = i + 1
	}
	return true
}

/**
	获取剩余地址数量
 */
func GetLeftAddress(xmysql *mysql.XMySQL) int64 {
	sqlstr := "select count(id) from address_pool where `user_id` = 0"
	row := xmysql.QueryRow(sqlstr)
	var count int64
	_ = row.Scan(&count)
	return count
}

func UseAddress(xmysql *mysql.XMySQL, userId int64) bool {
	coinsql := "select tokenname from coin where `status` = 1"
	rows, err := xmysql.Query(coinsql)
	if err != nil {
		fmt.Println("UseAddress select coin failed: ", err)
		return false
	}

	for rows.Next() {
		coinanme := ""
		_ = rows.Scan(&coinanme)
		CreateUserAddress(xmysql, userId, coinanme)
	}
	return true
}

func CreateUserAddress(xmysql *mysql.XMySQL, userId int64, coinanme string) {
	sql := "update address_pool set user_id = ? ,status = 1 where id in (select id FROM (SELECT `id` FROM `address_pool` WHERE status = 0 and coinname = ? LIMIT 1) as t)"
	_, err := xmysql.Exec(sql, userId, coinanme)

	if err != nil {
		fmt.Println("create address, roll back, %v", err)
	}

}

func GetUserIdByAddress(xmysql *mysql.XMySQL, coinname string, address string) int64 {
	//回去币种主链名称
	coinname = GetCoinTokenforName(xmysql, coinname)
	sql := "select user_id from address_pool where address = ? and coinname = ?"
	row := xmysql.QueryRow(sql, address, coinname)
	if row == nil {
		return 0
	}
	userId := int64(0)
	_ = row.Scan(&userId)

	return userId
}

func GetAddressByUserId(xmysql *mysql.XMySQL, coinname string, userId int64) string {
	sql := "select address from address_pool where user_id = ? and coinname = ?"

	row := xmysql.QueryRow(sql, userId, coinname)
	if row == nil {
		return ""
	}
	address := ""
	_ = row.Scan(&address)

	return address
}

func SelRessSum(xmysql *mysql.XMySQL, userId int64) int64 {
	sql := "SELECT COUNT(1) FROM address_pool "
	params := []interface{}{}
	if userId > 0 {
		sql = sql + " WHERE user_id = ? "
		params = append(params, userId)
	}
	sql += ""
	rows := xmysql.QueryRow(sql, params...)
	var num int64
	err := rows.Scan(&num)
	if err != nil {
		fmt.Println("查询用户收款地址Scan", err)
		log.Error("查询用户收款地址Scan", err)
	}
	return num
}

func SelRessIFUse(xmysql *mysql.XMySQL, coinName string, status int64) int64 {
	var num int64
	params := make([]interface{}, 0)
	params = append(params, status)
	sql := "SELECT COUNT(1) FROM address_pool WHERE `status` = ?"
	if coinName != "" {
		sql += " AND coinname = ? "
		params = append(params, coinName)
	}
	row := xmysql.QueryRow(sql, params...)
	err := row.Scan(&num)
	if err != nil {
		fmt.Println("查是地址否已使用", err)
		log.Error("查是地址否已使用", err)
		return num
	}
	return num
}

//获取临时数据
func GetAddressList(xmysql *mysql.XMySQL) ([]types.AddressPool,error) {

	sql := "SELECT id,address,status,coinname,user_id FROM address_pool order by id asc"
	rows,err := xmysql.Query(sql)
	if err != nil{
		return nil,err
	}

	var res = make([]types.AddressPool, 0)
	for rows.Next() {
		var addressPool types.AddressPool
		err = rows.Scan(&addressPool.Id, &addressPool.Address,&addressPool.Status,&addressPool.Coinname,&addressPool.UserId)
		if err != nil{
			return nil,err
		}
		res = append(res, addressPool)
	}

	return res,nil
}