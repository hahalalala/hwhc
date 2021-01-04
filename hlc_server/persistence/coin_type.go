package persistence

import (
	"encoding/json"
	"fmt"
	"github.com/EducationEKT/EKT/util"
	"github.com/hwhc/hlc_server/conf"
	"github.com/hwhc/hlc_server/mysql"
	"github.com/hwhc/hlc_server/types"
	"strconv"
)

const (
	HLC  = 1
	USDT = 2
	IDR  = 3
	VND  = 4
)

func GetCoinIdforName(xmysql *mysql.XMySQL, coinname string) int64 {
	coinsql := "select id from coin where `sortname` = ?"
	row := xmysql.QueryRow(coinsql, coinname)
	if row == nil {
		return 0
	}
	coinId := int64(0)
	_ = row.Scan(&coinId)

	return coinId
}

func GetCoinIdforTokenName(xmysql *mysql.XMySQL, coinname string) int64 {
	coinsql := "select id from coin where `tokenname` = ?"
	row := xmysql.QueryRow(coinsql, coinname)
	if row == nil {
		return 0
	}
	coinId := int64(0)
	_ = row.Scan(&coinId)

	return coinId
}

//获取主链币种名称
func GetCoinTokenfortokenname(xmysql *mysql.XMySQL, coinname string) string {
	coinsql := "select type from coin where `tokenname` = ?"
	row := xmysql.QueryRow(coinsql, coinname)
	if row == nil {
		return ""
	}
	tokenname := ""
	_ = row.Scan(&tokenname)

	return tokenname
}

//获取主链币种名称
func GetCoinTokenforName(xmysql *mysql.XMySQL, coinname string) string {
	coinsql := "select type from coin where `sortname` = ?"
	row := xmysql.QueryRow(coinsql, coinname)
	if row == nil {
		return ""
	}
	tokenname := ""
	_ = row.Scan(&tokenname)

	return tokenname
}

//获取可以提币的币种列表
func Getwithdrawal(xmysql *mysql.XMySQL) []types.CoinType {
	coinsql := "select id,coinname,sortname from coin where `app_withdrawal` = 1 or `withdrawal` = 1 "
	rows, err := xmysql.Query(coinsql)
	if err != nil {
		fmt.Println("GetCoins", err)
		return nil
	}
	list := make([]types.CoinType, 0)
	for rows.Next() {
		var record types.CoinType
		_ = rows.Scan(&record.Id, &record.Coinname, &record.Sortname)

		//price := GetRealTimePrice(xmysql, record.Id)
		//record.CoinQCPrice = price
		list = append(list, record)
	}

	return list
}

func GetCoinByCoinId(xmysql *mysql.XMySQL, id int64) types.CoinType {
	coinsql := "select id,sortname,coinname,contract_address,tokenname,type,recharge,withdrawal,chain_num,in_withdrawal,app_withdrawal,app_recharge ,`is_transfer` from coin where `id` = ? and `status` <> 0"
	row := xmysql.QueryRow(coinsql, id)

	var cointype types.CoinType

	err := row.Scan(&cointype.Id, &cointype.Sortname, &cointype.Coinname, &cointype.ContractCddress, &cointype.Tokenname, &cointype.Type, &cointype.Recharge, &cointype.Withdrawal, &cointype.ChainNum, &cointype.InWithdrawal, &cointype.AppWithdrawal, &cointype.AppRecharge, &cointype.IsTransfer)

	if err != nil {
		fmt.Println("GetCoinByCoinId failed, %v", err, "|CoinType", id)
	}
	return cointype
}

func SelCoins(xmysql *mysql.XMySQL, id int64) types.CoinType {
	coinsql := "select id,sortname,coinname,contract_address,tokenname,type,recharge,withdrawal,chain_num,in_withdrawal,app_withdrawal,app_recharge ,`is_transfer` from coin where `id` = ?"
	row := xmysql.QueryRow(coinsql, id)

	var cointype types.CoinType

	err := row.Scan(&cointype.Id, &cointype.Sortname, &cointype.Coinname, &cointype.ContractCddress, &cointype.Tokenname, &cointype.Type, &cointype.Recharge, &cointype.Withdrawal, &cointype.ChainNum, &cointype.InWithdrawal, &cointype.AppWithdrawal, &cointype.AppRecharge, &cointype.IsTransfer)

	if err != nil {
		fmt.Println("GetCoinByCoinId failed, %v", err, "|CoinType", id)
	}
	return cointype
}

func CoinListWallet(xmysql *mysql.XMySQL, useId int64) []types.CoinType {
	coinsql := "select is_transfer,app_withdrawal,app_recharge,c.id id , url , coinname,sortname,contract_address,tokenname,c.status status,recharge,withdrawal,in_withdrawal,user_id,amount,frozen_amount,tz_amount,chain_num,hz_amount from coin c  left JOIN (select * from user_amount where user_id = ?) ua on c.id = ua.type where c.status <> 0 ORDER BY chain_num asc"
	rows, err := xmysql.Query(coinsql, useId)
	if err != nil {
		fmt.Println("GetCoins", err)
		return nil
	}
	list := make([]types.CoinType, 0)
	for rows.Next() {
		var record types.CoinType
		_ = rows.Scan(&record.IsTransfer, &record.AppWithdrawal, &record.AppRecharge, &record.Id, &record.Url, &record.Coinname, &record.Sortname, &record.ContractCddress, &record.Tokenname, &record.Status, &record.Recharge, &record.Withdrawal, &record.InWithdrawal, &record.UserId, &record.Amount, &record.FrozenAmount, &record.TzAmount, &record.ChainNum, &record.HzAmount)

		//price := GetRealTimePrice(xmysql, record.Id)
		//if record.Id == USDT || record.Id == USDT_GCA || record.Id == USDT_YBK {
		//
		//	price = 1
		//}
		//
		//record.CoinQCPrice = price
		list = append(list, record)

	}

	return list
}

func GetRealTimePrice(xmysql *mysql.XMySQL, t int64) float64 {
	sql := "SELECT `price` FROM `price` WHERE `type` = ? ORDER BY `id` DESC LIMIT 1"
	row := xmysql.QueryRow(sql, t)
	price := float64(0)
	_ = row.Scan(&price)
	return price
}

func GetPriceByDate(xmysql *mysql.XMySQL, t int64,date string) float64 {
	sql := "SELECT `price` FROM `price` WHERE `type` = ? and date_str = ? ORDER BY `id` DESC LIMIT 1"
	row := xmysql.QueryRow(sql, t,date)
	price := float64(0)
	_ = row.Scan(&price)
	return price
}

func GetHLCPrice() (hlcPrice ,idrPrice ,vndPrice float64) {

	var url string
	if conf.GetConfig().Env == "test" {
		url = "http://18.166.67.216:8080/call/back/getCoinList"
	}else{
		url = "http://api.magnipay.com/call/back/getCoinList"
	}
	resp, err := util.HttpGet(url)
	if err != nil {
		fmt.Printf("GetHLCPrice err : %v \n",err)
	}
	var price []types.Hlc_price
	err = json.Unmarshal(resp, &price)
	for _, pre := range price {
		switch pre.NameCoin{
		case "MP":
			hlcPrice, _ = strconv.ParseFloat(pre.RateusdCoin, 64)
		case "IDR":
			idrPrice, _ = strconv.ParseFloat(pre.RateusdCoin, 64)
		case "VND":
			vndPrice, _ = strconv.ParseFloat(pre.RateusdCoin, 64)
		}
	}

	return hlcPrice,idrPrice,vndPrice
}

//@todo确认新加币种是否返回参与计算
func GetCoins(xmysql *mysql.XMySQL, useId int64) []types.CoinTypeReturn {
	coinsql := "select c.id id , url , coinname,sortname,tokenname,app_recharge,app_withdrawal,in_withdrawal,user_id,IFNULL(amount,0),IFNULL(frozen_amount,0),IFNULL(transfer_small_num,0),IFNULL(transfer_fee,0),transfer_big_num from coin c  left JOIN (select * from user_amount where user_id = ?) ua on c.id = ua.type where c.status <> 0 ORDER BY chain_num asc"
	rows, err := xmysql.Query(coinsql, useId)
	if err != nil {
		fmt.Println("GetCoins", err)
		return nil
	}
	list := make([]types.CoinTypeReturn, 0)
	hlcPrice,_,_:= GetHLCPrice()
	for rows.Next() {
		var record types.CoinTypeReturn
		_ = rows.Scan(&record.Id, &record.Url, &record.Coinname, &record.Sortname, &record.Tokenname, &record.AppRecharge, &record.AppWithdrawal, &record.InWithdrawal, &record.UserId, &record.Amount, &record.FrozenAmount, &record.TransferSmallNum, &record.TransferFee, &record.TransferBigNum)
		if record.Id == HLC {
			record.CoinQCPrice = record.Amount * hlcPrice
			record.FrozenUsdtPrice = record.FrozenAmount * hlcPrice
		} else if record.Id == USDT {
			record.CoinQCPrice = record.Amount * 1
			record.FrozenUsdtPrice = record.FrozenAmount * 1
		}
		list = append(list, record)
	}

	return list
}

func GetUserCoinAmount(xmysql *mysql.XMySQL, useId int64) []map[string]interface{} {
	coinsql := "select c.id id , url , coinname,sortname,contract_address,tokenname,c.status status,user_id,amount,frozen_amount,tz_amount,chain_num,recharge,withdrawal,in_withdrawal from coin c  left JOIN (select * from user_amount where user_id = ?) ua on c.id = ua.type where c.status <> 0"
	rows, err := xmysql.Query(coinsql, useId)
	if err != nil {
		fmt.Println("GetCoins", err)
		return nil
	}
	list := make([]map[string]interface{}, 0)
	for rows.Next() {
		var record types.CoinType
		_ = rows.Scan(&record.Id, &record.Url, &record.Coinname, &record.Sortname, &record.ContractCddress, &record.Tokenname, &record.Status, &record.UserId, &record.Amount, &record.FrozenAmount, &record.TzAmount, &record.ChainNum, &record.Recharge, &record.Withdrawal, &record.InWithdrawal)

		m := map[string]interface{}{
			record.Sortname: record.Amount,
		}

		list = append(list, m)
	}

	return list
}

func GetContactAddress(xmysql *mysql.XMySQL, coinid int64, parentCoinname string) string {
	coinsql := "SELECT contract_address from coin_contract_address where coinid = ? and parent_coinname = ?"
	row := xmysql.QueryRow(coinsql, coinid, parentCoinname)
	if row == nil {
		return ""
	}
	contactAddress := ""
	_ = row.Scan(&contactAddress)

	return contactAddress
}

func GetCoinContractName(xmysql *mysql.XMySQL, coinid int64) string {
	coinsql := "SELECT GROUP_CONCAT(parent_coinname) from coin_contract_address where coinid = ?"
	row := xmysql.QueryRow(coinsql, coinid)
	if row == nil {
		return ""
	}
	coinname := ""
	_ = row.Scan(&coinname)

	return coinname
}

//提币手续费
func Fee(xmysql *mysql.XMySQL, id int64) float64 {
	coinsql := "select transfer_fee from coin where `id` = ?"
	row := xmysql.QueryRow(coinsql, id)
	if row == nil {
		return 0
	}
	transferFee := float64(0)
	_ = row.Scan(&transferFee)

	return transferFee

}

//提币最大限额
func BigNum(xmysql *mysql.XMySQL, id int64) float64 {
	coinsql := "select IFNULL(`transfer_big_num`,'0') from coin where `id` = ?"
	row := xmysql.QueryRow(coinsql, id)
	if row == nil {
		return 0
	}
	bigNum := float64(0)
	_ = row.Scan(&bigNum)

	return bigNum

}

//提币最大限额
func Geturl(xmysql *mysql.XMySQL, id int64) string {
	coinsql := "select url from coin where `id` = ?"
	row := xmysql.QueryRow(coinsql, id)
	if row == nil {
		return ""
	}
	url := ""
	_ = row.Scan(&url)

	return url

}
