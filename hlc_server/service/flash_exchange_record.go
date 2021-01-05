package service

import (
	"fmt"
	"github.com/EducationEKT/xserver/x_err"
	"github.com/EducationEKT/xserver/x_http/x_resp"
	"github.com/hwhc/hlc_server/log"
	"github.com/hwhc/hlc_server/mysql"
	"github.com/hwhc/hlc_server/persistence"
	"github.com/hwhc/hlc_server/types"
)

func GetExchangeRecordList(userid int64) (int, []types.FlashExchangeRecord) {
	return persistence.GetExchangeRecordList(mysql.Get(), userid)
}

func GetExchangeList() []types.FlashExchange {
	return persistence.GetExchangeList(mysql.Get())
}

func AddExchangeRecord(exchange_id int64, num float64, userId int64, orderId string,hlc_price float64) (*x_resp.XRespContainer, *x_err.XErr) {
	xmysql := mysql.Get()

	if num <= 0 {
		return x_resp.Fail(-1094, "闪兑数量不能小于0", nil), nil
	}

	exchange := persistence.GetExchangeCoinid(xmysql, exchange_id) //查看闪对信息
	if exchange.BuyStatus != 1 {
		return x_resp.Fail(-1096, "当前交易已停止", nil), nil

	}

	//hlc_price := persistence.GetRealTimePrice(mysql.Get(), persistence.HLC)

	pay_price := persistence.GetRealTimePrice(mysql.Get(), exchange.PayCoinId)
	return_Price := persistence.GetRealTimePrice(mysql.Get(), exchange.ReturnCoinId)

	paycoins := persistence.GetCoinByCoinId(xmysql, exchange.PayCoinId)
	returncoins := persistence.GetCoinByCoinId(xmysql, exchange.ReturnCoinId)

	return_num := pay_price * num / return_Price //购买币种数量 * 购买币种价格 /  支付币种价格  =  需要支付币种数量

	exchange.Price = pay_price / return_Price //购买币种 / 支付币种 等于兑换价格

	wtCost := pay_price * num * exchange.Cost / hlc_price //手续费

	userAmount := persistence.GetUserAmount(xmysql, userId, exchange.PayCoinId) //查询支付币种余额
	if num > userAmount {
		fmt.Println("余额不足,可用余额：", userAmount, "--用户闪兑余额：", num+wtCost)
		return x_resp.Fail(-1003, "余额不足", nil), nil
	}
	//if exchange_id == 7 {
	//	total := persistence.GetTotalUserID(xmysql, exchange_id, userId) //用户已经闪兑的usdt 数量
	//	total = total + return_num
	//	if total > 5000 {
	//		return -1043
	//	}
	//}

	//是否限购
	if exchange.LimitStatus > 0 {
		purchased := persistence.Purchased(xmysql, exchange.ReturnCoinId) //已购买数量

		purchased = exchange.LimitTotal - (purchased + exchange.Purchased) //限购总额 - 已购买数量

		if purchased < num {
			return x_resp.Fail(-1093, "剩余已经购数量不足", nil), nil
		}
	}

	xmysqlBegin := mysql.Begin()
	defer xmysqlBegin.Commit()
	if persistence.ReduceUserAmount(xmysqlBegin, userId, persistence.HLC, wtCost) { //减手续费

		//手续费记录 finish
		if !(persistence.SaveTransfer(xmysqlBegin, userId, types.Fee, persistence.HLC, 0-wtCost, "", types.AMOUNT, orderId, 0, "", 1, 0,hlc_price) > 0) { //添加提现记录
			xmysqlBegin.Rollback()
			log.Error("Transfer 闪兑 扣取手续费 用户id：%s,币种类型 %s,扣取数量 %s", userId, persistence.HLC, wtCost)
			//	return  //转账失败
			return x_resp.Fail(-1011, "转账失败", nil), nil
		}

		if persistence.ReduceUserAmount(xmysqlBegin, userId, exchange.PayCoinId, num) { //减支付币种
			if persistence.AddExchangeRecord(xmysqlBegin, exchange.Id, num, return_num, exchange.Price, exchange.Name, userId, wtCost, paycoins.Sortname, returncoins.Sortname, exchange.Id) { //添加闪兑记录
				if persistence.AddUserAmount(xmysqlBegin, userId, exchange.ReturnCoinId, return_num, 0) { //增加闪兑获得的币种数量
					//添加支付币种记录 finish
					if !(persistence.SaveTransfer(xmysqlBegin, userId, types.FLASH_EXCHANGE, exchange.PayCoinId, 0-num, "", types.AMOUNT, orderId, wtCost, "", 1, 0,hlc_price) > 0) { //添加提现记录
						xmysqlBegin.Rollback()
						log.Error("Transfer 闪兑 支付币种添加数量到账用户id：%s,币种类型 %s,扣取数量 %s", userId, exchange.PayCoinId, num)
						//	return  //转账失败
						return x_resp.Fail(-1011, "转账失败", nil), nil
					}

					//添加支付币种记录 finish
					if !(persistence.SaveTransfer(xmysqlBegin, userId, types.FLASH_EXCHANGE, exchange.ReturnCoinId, return_num, "", types.AMOUNT, orderId, wtCost, "", 1, 0,hlc_price) > 0) { //添加提现记录
						xmysqlBegin.Rollback()
						log.Error("Transfer 闪兑 支付币种添加数量到账用户id：%s,币种类型 %s,扣取数量 %s", userId, exchange.ReturnCoinId, num)
						//	return  //转账失败
						return x_resp.Fail(-1011, "转账失败", nil), nil
					}

					return x_resp.Success(1), nil
				}
				fmt.Println("增加闪兑获得的币种数量 sql err ")
			} else {
				log.Error("余额不足,用户ID ： ", userId, "-需要支付余额：", num, "-闪对类型：", exchange.Id)
				xmysqlBegin.Rollback()
				return x_resp.Fail(-1002, "余额不足", nil), nil
			}
		} else {
			log.Error("余额不足,用户ID ： ", userId, "-需要支付余额：", num, "-闪对类型：", exchange.Id)
			xmysqlBegin.Rollback()
			return x_resp.Fail(-1003, "余额不足", nil), nil
		}
	} else {
		log.Error("手续费不足，需要扣除", wtCost, "用户ID：", userId)
		xmysqlBegin.Rollback()
		return x_resp.Fail(-12, "手续费余额不足", nil), nil
	}

	fmt.Println("减闪兑所需要的wt sql err ")
	xmysqlBegin.Rollback()
	return x_resp.Fail(-1002, "交易异常", nil), nil
}

func ExchangeBuyAmount(userId, exchangeId int64) float64 {
	xmysql := mysql.Get()
	level := persistence.SelUserLevel(xmysql, userId)
	purch := persistence.PurchasedUserID(xmysql, exchangeId, userId)
	if level == 0 {
		return 0
	} else if level == 1 {
		return 10000 - purch
	} else if level == 2 {
		return 20000 - purch
	} else if level == 3 {
		return 30000 - purch
	} else if level == 4 {
		return 40000 - purch
	} else if level == 5 {
		return 50000 - purch
	} else if level == 6 {
		return 60000 - purch
	} else {
		return 0
	}
}
