package service

import (
	"fmt"
	"github.com/hwhc/hlc_server/mysql"
	"github.com/hwhc/hlc_server/persistence"
	"github.com/hwhc/hlc_server/types"
	"strconv"
	"strings"
)

func Recharge(address string, amount float64, txHash string, coinname string) {
	fmt.Print("查找地址..  %s", coinname, "  ", address)
	coinid := persistence.GetCoinIdforTokenName(mysql.Get(), coinname)
	fmt.Println("coinid", coinid)
	var userId int64
	if strings.ToLower(coinname) == strings.ToLower("EOS") || strings.ToLower(coinname) == strings.ToLower("XRP") {
		fmt.Println("--xrp 或者 eos 充值")
		userId, _ = strconv.ParseInt(address, 10, 64)
	} else {
		coininfo := persistence.GetCoinByCoinId(mysql.Get(), coinid)
		fmt.Println("coininfo.Type", coininfo.Type)
		userId = GetUserIdByAddress(coininfo.Sortname, address)
	}
	fmt.Println("userId", userId)

	hlcPrice := persistence.GetRealTimePrice(mysql.Get(),persistence.HLC)

	if userId > 0 {
		fmt.Print("---找到  %s", userId, "  ", address)
		xmysql := mysql.Begin()
		defer xmysql.Commit()
		if persistence.SaveTransfer(xmysql, userId, types.RECHARGE, coinid, amount, address, types.AMOUNT, txHash, 0, "", 1, 0,hlcPrice) > 0 {
			//coinid := persistence.GetCoinIdforName(xmysql, coinname) //查询币种类型ID、
			if coinid > 0 {
				if !persistence.AddUserAmount(xmysql, userId, coinid, amount, 0) {
					xmysql.Rollback()
				}
			} else {
				xmysql.Rollback()
			}

		} else {
			xmysql.Rollback()
		}
	}
}
