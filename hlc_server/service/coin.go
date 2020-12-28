package service

import (
	"fmt"
	"github.com/EducationEKT/xserver/x_err"
	"github.com/EducationEKT/xserver/x_http/x_resp"
	"github.com/EducationEKT/xserver/x_utils/x_random"
	"github.com/hwhc/hlc_server/log"
	"github.com/hwhc/hlc_server/mysql"
	"github.com/hwhc/hlc_server/persistence"
	"github.com/hwhc/hlc_server/types"
	"github.com/hwhc/hlc_server/util"
	"github.com/shopspring/decimal"
	"math"
	"strconv"
	"time"
)

func GetFuTou() (*x_resp.XRespContainer, *x_err.XErr) {

	beginTime := "2020-11-17 00:00:00"
	endTime := "2020-11-17 23:59:59"

	list := persistence.GetFuTou(mysql.Get(), beginTime, endTime)

	var personNum int64       //用户总数量
	var receiverTotal float64 //接受转账总金额
	var constTotal float64    //转出去总金额
	var realTotal float64     //实际要统计的总金额 ,去接受转账总金额、转出去总金额两个对比的最小值

	for _, m := range list {
		for userId, AmountTotal := range m {
			personNum += 1
			receiverTotal += AmountTotal

			constNum := persistence.GetReceiveTotal(mysql.Get(), userId, beginTime, endTime)
			absConstNum := math.Abs(constNum)

			constTotal += absConstNum

			if absConstNum < AmountTotal {
				realTotal += absConstNum
			} else {
				realTotal += AmountTotal
			}
		}

	}

	fmt.Printf("用户总数量 : %d \n", personNum)
	fmt.Printf("接受转账总金额 : %.8f \n", receiverTotal)
	fmt.Printf("转出去总金额 : %.8f \n", constTotal)
	fmt.Printf("实际统计总金额 : %.8f \n", realTotal)

	return nil, nil
}

//获取交易
func GetTransfer(txHash string, typ int64) (*x_resp.XRespContainer, *x_err.XErr) {

	if typ == types.SHOP {
		//消费之前历史原因tx_hash写死了拼了一个_1
		txHash = txHash + "_1"
	}
	transfer := persistence.GetTransfer(mysql.Get(), txHash, typ)
	return x_resp.Success(transfer), nil
}

//扣除系统脚本充值 "一定时间 : spaceTime" 未使用的金额
func ReduceFreeRecharge(coinId, spaceTime int64) (*x_resp.XRespContainer, *x_err.XErr) {

	userAmountList := persistence.GetHasPriceUserAmountList(mysql.Get(), coinId)
	if len(userAmountList) > 0 {
		for _, userAmount := range userAmountList {
			lastTime := persistence.GetFreeRechargeLastTime(mysql.Get(), userAmount.UserId, coinId)
			if len(lastTime) == 0 {
				continue
			}
			loc, _ := time.LoadLocation("Asia/Shanghai")        //设置时区
			tm, err := time.ParseInLocation("2006-01-02 15:04:05", lastTime, loc)
			if err != nil {
				log.Error("ReduceFreeRecharge Parse lastTime err : %v ,userid:%d,coinId:%d,lastTime:%s", err, userAmount.UserId, coinId, lastTime)
				continue
			}
			lastTimeTimestamp := tm.Unix() //获取最后一条记录时间
			finalLastTimeTimestamp :=  lastTimeTimestamp + spaceTime //加上这个间隔时间
			now := time.Now().Unix()
			if finalLastTimeTimestamp <= now {
				//超过spaceTime间隔时间了,扣除账户所有
				orderId := fmt.Sprintf("扣除未使用的系统充值额度-%d-%s", userAmount.UserId, util.Krand(8, util.KC_RAND_KIND_ALL))
				_, err := ReAmount(userAmount.UserId, userAmount.Amount, orderId, types.REDUCE_SYSTEM_SCRIPT_FREE_RECHARGE, coinId, 0)
				if err != nil {
					log.Error("扣除未使用的系统充值额度 失败 userId:%d,amount:%.8f")
				} else {
					log.Info("扣除成功 coinid:%d , userid:%d ,amount:%f", coinId, userAmount.UserId, userAmount.Amount)
				}
			} else {
				log.Info("未到时间 coinid:%d , userid:%d ,amount:%f", coinId, userAmount.UserId, userAmount.Amount)
			}
		}
	}

	return x_resp.Success(0), nil
}

//提现列表
func TakeRecording(page, size int64, is_shop int64) ([]types.Transfer, int64) {
	start := (page - 1) * size
	end := start + size
	transfers := persistence.TakeRecording(mysql.Get(), start, end, is_shop)
	count := persistence.TakeRecordingNum(mysql.Get(), is_shop)
	return transfers, count
}

func TransferGoBack(txDesc string, userId, id int64) int {
	xmysql := mysql.Begin()
	defer xmysql.Commit()

	transfer := persistence.SelIdTransfer(xmysql, id)

	if !persistence.UpTransferStatus(xmysql, txDesc, userId, transfer.Tx_hash, 0, -1) {
		xmysql.Rollback()
		return -100020
	}

	if !persistence.AddUserAmount(xmysql, userId, transfer.CoinId, 0-transfer.Amount, 0) {
		xmysql.Rollback()
		return -100021
	}

	if transfer.IsShop > 0 {
		if !persistence.AddUserAmount(xmysql, userId, persistence.USDT, transfer.Fee, 0) {
			xmysql.Rollback()
			return -100022
		}
	} else {
		if !persistence.AddUserAmount(xmysql, userId, persistence.HLC, transfer.Fee, 0) {
			xmysql.Rollback()
			return -100022
		}
	}

	return 0
}

func GetUserAmount(userId, t int64) float64 {
	return persistence.GetUserAmount(mysql.Get(), userId, t)
}

func GetUserFrozenAmount(userId, t int64) (float64, float64) {
	return persistence.GetUserFrozenAmount(mysql.Get(), userId, t)
}

func GetTransferList(userId, coin_id int64, size, types, lastId int64) ([]map[string]interface{}, int64) {
	cout := persistence.TransferCount(mysql.Get(), userId, coin_id, types)
	return persistence.TransferList(mysql.Get(), userId, coin_id, size, types, lastId), cout
}

func TransferbyId(userId, transferId int64, cid int64) types.Transfer {
	return persistence.TransferbyId(mysql.Get(), userId, transferId, cid)

}

func Transfer_wchat(orderId string, cid_int int64, address string, amount float64, userId int64, memo string, is_shop int64) (*x_resp.XRespContainer, *x_err.XErr) {

	coininfo := persistence.GetCoinByCoinId(mysql.Get(), cid_int)

	if coininfo.Sortname == "" {
		return x_resp.Fail(-1056, "币种id传输错误", nil), nil
	}

	useramount := persistence.GetUserAmount(mysql.Get(), userId, cid_int)
	userwtamount := persistence.GetUserAmount(mysql.Get(), userId, persistence.HLC)
	hlcPrice := persistence.GetRealTimePrice(mysql.Get(), persistence.HLC)   //0.3
	usdtPrice := persistence.GetRealTimePrice(mysql.Get(), persistence.USDT) //1
	coinPrice := persistence.GetRealTimePrice(mysql.Get(), cid_int)          //1
	transferFee:= Fee(cid_int) //写死的0.05改为动态的

	//wtFee := amount * transferFee * coinPrice / hlcPrice
	rate:= decimal.NewFromFloat(transferFee).Mul(decimal.NewFromFloat(coinPrice)).Div(decimal.NewFromFloat(hlcPrice)).Truncate(8)
	wtFee,exact :=rate.Mul(decimal.NewFromFloat(amount)).Float64()
	if !exact {
		log.Error(fmt.Sprintf("Transfer_wchat decimal fail orderId : %s ,cid_int:%d ,address:%s,amount:%.8f,userId:%d,isShop:%d,transferFee:%.8f,coinPrice:%.8f,hlcPrice:%.8f",
			orderId, cid_int, address,amount,userId,is_shop,transferFee,coinPrice,hlcPrice))
		return x_resp.Fail(-1057, "[1]手续费余额不足", nil), nil
	}

	fee_coin := persistence.HLC
	if is_shop > 0 {

		//wtFee = amount * 0.01 * coinPrice / usdtPrice

		shopRate := decimal.NewFromFloat(0.01).Mul(decimal.NewFromFloat(coinPrice)).Div(decimal.NewFromFloat(usdtPrice)).Truncate(8)
		wtFee ,exact = decimal.NewFromFloat(amount).Mul(shopRate).Float64()
		if !exact{
			log.Error(fmt.Sprintf("[2]Transfer_wchat decimal fail orderId : %s ,cid_int:%d ,address:%s,amount:%.8f,userId:%d,isShop:%d,transferFee:%.8f,coinPrice:%.8f,hlcPrice:%.8f",
				orderId, cid_int, address,amount,userId,is_shop,0.01,coinPrice,hlcPrice))
			return x_resp.Fail(-1058, "[2]手续费余额不足", nil), nil
		}
		userwtamount = persistence.GetUserAmount(mysql.Get(), userId, persistence.USDT)
		fee_coin = persistence.USDT
	} else {
		if wtFee < 100 {
			wtFee = 100
		}
	}

	wtFee,exact = decimal.NewFromFloat(wtFee).Truncate(5).Float64() //保留5位
	if !exact {
		log.Error(fmt.Sprintf("[3]Transfer_wchat decimal fail orderId : %s ,cid_int:%d ,address:%s,amount:%.8f,userId:%d,isShop:%d,transferFee:%.8f,coinPrice:%.8f,hlcPrice:%.8f ,wtFee:%.8f",
			orderId, cid_int, address,amount,userId,is_shop,0.01,coinPrice,hlcPrice,wtFee))
		return x_resp.Fail(-1059, "[3]手续费余额不足", nil), nil
	}

	if decimal.NewFromFloat(userwtamount).LessThan(decimal.NewFromFloat(wtFee)) { //userwtamount < wtFee
		log.Error("Transfer Transfer_wchat 扣取余额失败 用户手续费余额不足 用户id：%s,币种类型 %s,扣取数量 %s", userId, cid_int, amount)
		return x_resp.Fail(-12, "手续费余额不足", nil), nil
	}
	if decimal.NewFromFloat(useramount).LessThan(decimal.NewFromFloat(amount)) { // useramount < amount
		log.Error("Transfer_wchat 查询用户余额不足,用户ID:%s,币种：%s,余额：%s", userId, cid_int, useramount)
		fmt.Println("Transfer_wchat 查询用户余额不足,用户ID:%s,币种：%s,余额：%s", userId, cid_int, useramount)
		return x_resp.Fail(-1003, "账户余额不足", nil), nil
	}


	xmysql := mysql.Begin()
	defer xmysql.Commit()
	if !(persistence.SaveTransfer(xmysql, userId, types.TRANSFER, cid_int, 0-amount, address, types.AMOUNT, orderId, wtFee, "", 0, is_shop) > 0) { //添加提现记录
		xmysql.Rollback()
		log.Error("Transfer Transfer_wchat 保存用户提现信息 到账用户id：%s,币种类型 %s,扣取数量 %s", userId, cid_int, amount)
		return x_resp.Fail(-1014, "转账失败", nil), nil
	}

	//减掉用户余额
	if !persistence.ReduceUserAmount(xmysql, userId, cid_int, amount) {
		log.Error("Transfer_wchat 减用户资产出现错误,用户ID:%s,币种：%s,余额：%s", userId, cid_int, useramount)
		fmt.Println("Transfer_wchat 减用户资产出现错误,用户ID:%s,币种：%s,余额：%s", userId, cid_int, useramount)
		xmysql.Rollback()
		return x_resp.Fail(-1007, "余额不足", nil), nil
	}

	//手续费记录
	if !(persistence.SaveTransfer(xmysql, userId, types.Fee, int64(fee_coin), 0-wtFee, "", types.AMOUNT, orderId, 0, "", 0, is_shop) > 0) { //添加提现记录
		xmysql.Rollback()
		log.Error("Transfer 内部转账 扣取手续费 到账用户id：%s,币种类型 %s,扣取数量 %s", userId, cid_int, amount)
		//	return  //转账失败
		return x_resp.Fail(-1011, "转账失败", nil), nil
	}
	//减手续费可用
	if !persistence.ReduceUserAmount(xmysql, userId, int64(fee_coin), wtFee) {
		xmysql.Rollback()
		log.Error("Transfer_wchat 减手续费可用 扣取余额失败 用户id：%s,币种类型 %s,扣取数量 %s", userId, cid_int, amount)
		return x_resp.Fail(-1003, "手续费余额不足", nil), nil
	}

	return x_resp.Success(0), nil
}

func Transfer_otc(orderId string, cid_int int64, amount float64, userId int64, memo string, is_shop int64) (*x_resp.XRespContainer, *x_err.XErr) {

	coininfo := persistence.GetCoinByCoinId(mysql.Get(), cid_int)

	if coininfo.Sortname == "" {
		return x_resp.Fail(-1056, "币种id传输错误", nil), nil
	}

	xmysql := mysql.Begin()
	defer xmysql.Commit()
	// 验证余额    验证余额    验证余额

	useramount := persistence.GetUserAmount(mysql.Get(), userId, cid_int)
	userwtamount := persistence.GetUserAmount(mysql.Get(), userId, persistence.HLC)

	hlcPrice := persistence.GetRealTimePrice(mysql.Get(), persistence.HLC)
	//usdtPrice := persistence.GetRealTimePrice(mysql.Get(), persistence.USDT)
	coinPrice := persistence.GetRealTimePrice(mysql.Get(), cid_int)
	wtFee := amount * 0.05 * coinPrice / hlcPrice
	fee_coin := persistence.HLC

	if userwtamount < wtFee {
		log.Error("Transfer Transfer_wchat 扣取余额失败 用户手续费余额不足 用户id：%s,币种类型 %s,扣取数量 %s", userId, cid_int, amount)
		xmysql.Rollback()
		return x_resp.Fail(-12, "手续费余额不足", nil), nil
	}

	if useramount < amount {
		log.Error("Transfer_wchat 查询用户余额不足,用户ID:%s,币种：%s,余额：%s", userId, cid_int, useramount)
		fmt.Println("Transfer_wchat 查询用户余额不足,用户ID:%s,币种：%s,余额：%s", userId, cid_int, useramount)
		xmysql.Rollback()
		return x_resp.Fail(-1003, "账户余额不足", nil), nil
	}

	if !(persistence.SaveTransfer(xmysql, userId, types.OTC_TRANSFER, cid_int, 0-amount, "", types.AMOUNT, orderId, wtFee, "", 0, is_shop) > 0) { //添加提现记录
		xmysql.Rollback()
		log.Error("Transfer Transfer_wchat 保存用户提现信息 到账用户id：%s,币种类型 %s,扣取数量 %s", userId, cid_int, amount)
		return x_resp.Fail(-1014, "转账失败", nil), nil
	}

	//减掉用户余额
	if !persistence.ReduceUserAmount(xmysql, userId, cid_int, amount) {
		log.Error("Transfer_wchat 减用户资产出现错误,用户ID:%s,币种：%s,余额：%s", userId, cid_int, useramount)
		fmt.Println("Transfer_wchat 减用户资产出现错误,用户ID:%s,币种：%s,余额：%s", userId, cid_int, useramount)
		xmysql.Rollback()
		return x_resp.Fail(-1007, "余额不足", nil), nil
	}

	//手续费记录
	if !(persistence.SaveTransfer(xmysql, userId, types.Fee, int64(fee_coin), 0-wtFee, "", types.AMOUNT, orderId, 0, "", 0, is_shop) > 0) { //添加提现记录
		xmysql.Rollback()
		log.Error("Transfer 内部转账 扣取手续费 到账用户id：%s,币种类型 %s,扣取数量 %s", userId, cid_int, amount)
		//	return  //转账失败
		return x_resp.Fail(-1011, "转账失败", nil), nil
	}
	//减手续费可用
	if !persistence.ReduceUserAmount(xmysql, userId, int64(fee_coin), wtFee) {
		xmysql.Rollback()
		log.Error("Transfer_wchat 减手续费可用 扣取余额失败 用户id：%s,币种类型 %s,扣取数量 %s", userId, cid_int, amount)
		return x_resp.Fail(-1003, "手续费余额不足", nil), nil
	}

	return x_resp.Success(0), nil
}

func Transfer(userId, coin_id int64, amount float64, re_userId int64, sortname string) (*x_resp.XRespContainer, *x_err.XErr) {

	coininfo := persistence.GetCoinByCoinId(mysql.Get(), coin_id)

	if coininfo.Sortname == "" {
		return x_resp.Fail(-1056, "币种id传输错误", nil), nil
	}

	xmysql := mysql.Begin()
	defer xmysql.Commit()

	if amount <= 0 {
		return x_resp.Fail(-1000, "转账金额有误", nil), nil
	}
	hlcPrice := 1.1
	coinprice := 1.0

	wtRate := 0.0 // 0.003
	wtFee := amount * wtRate * coinprice / hlcPrice

	userhlcamount := persistence.GetUserAmount(mysql.Get(), userId, persistence.HLC)

	if userhlcamount < wtFee {
		log.Error("Transfer 内部转账 扣取余额失败 用户手续费余额不足 用户id：%s,币种类型 %s,扣取数量 %s", userId, coin_id, amount)
		return x_resp.Fail(-12, "手续费余额不足", nil), nil
	}
	intId := strconv.FormatInt(userId, 10)
	orderId := fmt.Sprintf("prod_%s_%d_%s", intId, time.Now().UnixNano(), x_random.RandomString(5))

	if wtRate > 0 {
		//减手续费可用
		if !persistence.ReduceUserAmount(xmysql, userId, persistence.HLC, wtFee) {
			xmysql.Rollback()
			log.Error("Transfer 内部转账 扣取余额失败 用户id：%s,币种类型 %s,扣取数量 %s", userId, coin_id, amount)
		}

		//手续费记录
		if !(persistence.SaveTransfer(xmysql, userId, types.Fee, coin_id, 0-wtFee, "", types.AMOUNT, orderId, 0, "", 1, 0) > 0) { //添加提现记录
			xmysql.Rollback()
			log.Error("Transfer 内部转账 扣取手续费 到账用户id：%s,币种类型 %s,扣取数量 %s", re_userId, coin_id, amount)
			//	return  //转账失败
			return x_resp.Fail(-1011, "转账失败", nil), nil
		}
	}
	//减可用
	if !persistence.ReduceUserAmount(xmysql, userId, coin_id, amount) {
		xmysql.Rollback()
		log.Error("Transfer 内部转账 扣取余额失败 用户id：%s,币种类型 %s,扣取数量 %s", userId, coin_id, amount)
		//return -1003 //
		return x_resp.Fail(-1003, "余额不足", nil), nil
	}

	//内部转帐	验证手机好
	if !persistence.AddUserAmount(xmysql, re_userId, coin_id, amount, 0) {
		xmysql.Rollback()
		log.Error("Transfer 内部转账 添加用户资产失败 到账用户id：%s,币种类型 %s,扣取数量 %s", re_userId, coin_id, amount)
		return x_resp.Fail(-1016, "转账失败", nil), nil
	}
	//存储内部转账信息_出账
	if !(persistence.SaveTransfer(xmysql, userId, types.IN_TRANSFER, coin_id, 0-amount, strconv.FormatInt(re_userId, 10), types.AMOUNT, orderId, wtFee, "", 1, 0) > 0) { //添加提现记录
		xmysql.Rollback()
		log.Error("Transfer 内部转账 保存用户提现信息 到账用户id：%s,币种类型 %s,扣取数量 %s", re_userId, coin_id, amount)
		return x_resp.Fail(-1014, "转账失败", nil), nil
	}

	//存储内部转账信息_入账
	if persistence.SaveTransfer(xmysql, re_userId, types.IN_TRANSFER, coin_id, amount, "", types.AMOUNT, orderId, wtFee, "", 1, 0) > 0 {
		return x_resp.Success(0), nil
	} else {
		xmysql.Rollback()
		log.Error("Transfer 内部转账 RechargeRecord 添加充值记录失败 到账用户id：%s,币种类型 %s,扣取数量 %s", re_userId, coin_id, amount)
		return x_resp.Fail(-1015, "转账失败", nil), nil
	}

}

func AddAmount(userid int64, amount float64, orderId string, typess int64, coinId int64, is_shop int64) (*x_resp.XRespContainer, *x_err.XErr) {

	coininfo := persistence.GetCoinByCoinId(mysql.Get(), coinId)

	if coininfo.Sortname == "" {
		return x_resp.Fail(-1056, "币种id传输错误", nil), nil
	}

	xmysql := mysql.Begin()
	defer xmysql.Commit()
	//
	//if typess != 7 && typess != 6 && typess != 9 && typess != 10 &&
	//	typess != 12 && typess != 13 && typess != 14 && typess != 15 && typess != 16 && typess != 17 && typess != 18 && typess != 19 {
	//	return x_resp.Fail(-1023, "业务不符合", nil), nil
	//}

	//内部转帐	验证手机好
	if !persistence.AddUserAmount(xmysql, userid, coinId, amount, is_shop) {
		xmysql.Rollback()
		log.Error("Transfer AddAmount 添加用户资产失败 到账用户id：%s,币种类型 %s,扣取数量 %s", userid, coinId, amount)
		return x_resp.Fail(-1016, "转账失败", nil), nil
	}
	//存储内部转账信息_出账
	if !(persistence.SaveTransfer(xmysql, userid, typess, coinId, amount, "", types.AMOUNT, orderId, 0, "", 1, 0) > 0) { //添加提现记录
		xmysql.Rollback()
		log.Error("Transfer AddAmount 保存用户提现信息 到账用户id：%s,币种类型 %s,扣取数量 %s", userid, coinId, amount)
		return x_resp.Fail(-1014, "转账失败", nil), nil
	}
	return x_resp.Success(0), nil
}

func ReAmount(userid int64, amount float64, orderId string, typess int64, coinId int64, is_shop int64) (*x_resp.XRespContainer, *x_err.XErr) {

	coininfo := persistence.GetCoinByCoinId(mysql.Get(), coinId)

	if coininfo.Sortname == "" {
		return x_resp.Fail(-1056, "币种id传输错误", nil), nil
	}

	xmysql := mysql.Begin()
	defer xmysql.Commit()

	//减可用
	if !persistence.ReduceUserAmount(xmysql, userid, coinId, amount) {
		xmysql.Rollback()
		log.Error("Transfer ReAmount 添加用户资产失败 到账用户id：%s,币种类型 %s,扣取数量 %s", userid, coinId, amount)
		return x_resp.Fail(-1003, "余额不足", nil), nil
	}
	//存储内部转账信息_出账
	if !(persistence.SaveTransfer(xmysql, userid, typess, coinId, 0-amount, "", types.AMOUNT, orderId+"_1", 0, "", 1, is_shop) > 0) { //添加提现记录
		xmysql.Rollback()
		log.Error("Transfer ReAmount 保存用户提现信息 到账用户id：%s,币种类型 %s,扣取数量 %s", userid, coinId, amount)
		return x_resp.Fail(-1014, "转账失败", nil), nil
	}

	return x_resp.Success(0), nil
}

func AddFrozenToAmount(userid int64, amount float64, orderId string, typess int64, coinId int64, is_shop int64) (*x_resp.XRespContainer, *x_err.XErr) {

	coininfo := persistence.GetCoinByCoinId(mysql.Get(), coinId)

	if coininfo.Sortname == "" {
		return x_resp.Fail(-1056, "币种id传输错误", nil), nil
	}

	xmysql := mysql.Begin()
	defer xmysql.Commit()

	//加可用
	if !persistence.AddUserFrozenAmount(xmysql, userid, coinId, amount, is_shop) {
		xmysql.Rollback()
		log.Error("Transfer AddFrozenToAmount 添加用户资产失败 到账用户id：%s,币种类型 %s,扣取数量 %s", userid, coinId, amount)
		return x_resp.Fail(-1016, "转账失败", nil), nil
	}
	//存储内部转账信息_出账
	if !(persistence.SaveTransfer(xmysql, userid, typess, coinId, amount, "", types.FROZEN_AMOUNT, orderId, 0, "", 1, 0) > 0) { //添加提现记录
		xmysql.Rollback()
		log.Error("Transfer AddFrozenToAmount 保存用户提现信息 到账用户id：%s,币种类型 %s,扣取数量 %s", userid, coinId, amount)
		return x_resp.Fail(-1014, "转账失败", nil), nil
	}

	//减冻结
	if !persistence.ReduceUserAmount(xmysql, userid, coinId, amount) {
		xmysql.Rollback()
		log.Error("Transfer AddFrozenToAmount 添加用户资产失败 到账用户id：%s,币种类型 %s,扣取数量 %s", userid, coinId, amount)
		return x_resp.Fail(-1003, "余额不足", nil), nil
	}
	//存储内部转账信息_出账
	if !(persistence.SaveTransfer(xmysql, userid, typess, coinId, 0-amount, "", types.AMOUNT, orderId+"_1", 0, "", 1, 0) > 0) { //添加提现记录
		xmysql.Rollback()
		log.Error("Transfer AddFrozenToAmount 保存用户提现信息 到账用户id：%s,币种类型 %s,扣取数量 %s", userid, coinId, amount)
		return x_resp.Fail(-1014, "转账失败", nil), nil
	}

	return x_resp.Success(0), nil
}

func ShopTranfer(userid int64, amount float64, orderId string, typess int64, coinId int64, shop_user_id int64, shop_amout float64) (*x_resp.XRespContainer, *x_err.XErr) {

	coininfo := persistence.GetCoinByCoinId(mysql.Get(), coinId)

	if coininfo.Sortname == "" {
		return x_resp.Fail(-1056, "币种id传输错误", nil), nil
	}

	xmysql := mysql.Begin()
	defer xmysql.Commit()

	//加可用
	if !persistence.AddUserAmount(xmysql, shop_user_id, persistence.USDT, shop_amout, 1) {
		xmysql.Rollback()
		log.Error("Transfer AddFrozen 添加用户资产失败 到账用户id：%s,币种类型 %s,扣取数量 %s", shop_user_id, persistence.USDT, shop_amout)
		return x_resp.Fail(-1016, "转账失败", nil), nil
	}
	//存储内部转账信息_出账
	if !(persistence.SaveTransfer(xmysql, shop_user_id, typess, persistence.USDT, shop_amout, "", types.AMOUNT, orderId, 0, "", 1, 1) > 0) { //添加提现记录
		xmysql.Rollback()
		log.Error("Transfer AddFrozen 保存用户提现信息 到账用户id：%s,币种类型 %s,扣取数量 %s", shop_user_id, persistence.USDT, shop_amout)
		return x_resp.Fail(-1014, "转账失败", nil), nil
	}

	//减冻结
	if !persistence.ReduceUserAmount(xmysql, userid, coinId, amount) {
		xmysql.Rollback()
		log.Error("Transfer AddFrozen 添加用户资产失败 到账用户id：%s,币种类型 %s,扣取数量 %s", userid, coinId, amount)
		return x_resp.Fail(-1003, "余额不足", nil), nil
	}
	//存储内部转账信息_出账
	if !(persistence.SaveTransfer(xmysql, userid, typess, coinId, 0-amount, "", types.AMOUNT, orderId+"_1", 0, "", 1, 0) > 0) { //添加提现记录
		xmysql.Rollback()
		log.Error("Transfer AddFrozen 保存用户提现信息 到账用户id：%s,币种类型 %s,扣取数量 %s", userid, coinId, amount)
		return x_resp.Fail(-1014, "转账失败", nil), nil
	}

	return x_resp.Success(0), nil
}

func MergeCoinns(userid int64, amount float64, orderId string, typess int64, coinId int64, is_shop int64, re_coin_id int64, re_amount float64) (*x_resp.XRespContainer, *x_err.XErr) {

	coininfo := persistence.GetCoinByCoinId(mysql.Get(), coinId)

	if coininfo.Sortname == "" {
		return x_resp.Fail(-1056, "币种id传输错误", nil), nil
	}

	xmysql := mysql.Begin()
	defer xmysql.Commit()

	if amount > 0 {
		//加可用
		if !persistence.ReduceUserAmount(xmysql, userid, coinId, amount) {
			xmysql.Rollback()
			log.Error("Transfer MergeCoinns 减用户资产失败 到账用户id：%s,币种类型 %s,扣取数量 %s", userid, coinId, amount)
			return x_resp.Fail(-1016, "余额不足", nil), nil
		}
		//存储内部转账信息_出账
		if !(persistence.SaveTransfer(xmysql, userid, typess, coinId, 0-amount, "", types.AMOUNT, orderId, 0, "", 1, 0) > 0) { //添加提现记录
			xmysql.Rollback()
			log.Error("Transfer AddFrozen 保存用户提现信息 到账用户id：%s,币种类型 %s,扣取数量 %s", userid, coinId, amount)
			return x_resp.Fail(-1014, "转账失败", nil), nil
		}
	}

	if re_amount > 0 {
		//减冻结
		if !persistence.ReduceUserAmount(xmysql, userid, re_coin_id, re_amount) {
			xmysql.Rollback()
			log.Error("Transfer MergeCoinns _2  添加用户资产失败 到账用户id：%s,币种类型 %s,扣取数量 %s", userid, re_coin_id, re_amount)
			return x_resp.Fail(-1003, "余额不足", nil), nil
		}
		//存储内部转账信息_出账
		if !(persistence.SaveTransfer(xmysql, userid, typess, re_coin_id, 0-re_amount, "", types.AMOUNT, orderId+"_1", 0, "", 1, 0) > 0) { //添加提现记录
			xmysql.Rollback()
			log.Error("Transfer AddFrozen 保存用户提现信息 到账用户id：%s,币种类型 %s,扣取数量 %s", userid, coinId, amount)
			return x_resp.Fail(-1014, "转账失败", nil), nil
		}
	}

	return x_resp.Success(0), nil
}

func AddAmounToFrozen(userid int64, amount float64, orderId string, typess int64, coinId int64, is_shop int64) (*x_resp.XRespContainer, *x_err.XErr) {

	coininfo := persistence.GetCoinByCoinId(mysql.Get(), coinId)

	if coininfo.Sortname == "" {
		return x_resp.Fail(-1056, "币种id传输错误", nil), nil
	}

	xmysql := mysql.Begin()
	defer xmysql.Commit()

	//加可用
	if !persistence.AddUserAmount(xmysql, userid, coinId, amount, is_shop) {
		xmysql.Rollback()
		log.Error("Transfer AddFrozen 添加用户资产失败 到账用户id：%s,币种类型 %s,扣取数量 %s", userid, coinId, amount)
		return x_resp.Fail(-1016, "转账失败", nil), nil
	}
	//存储内部转账信息_出账
	if !(persistence.SaveTransfer(xmysql, userid, typess, coinId, amount, "", types.AMOUNT, orderId, 0, "", 1, 0) > 0) { //添加提现记录
		xmysql.Rollback()
		log.Error("Transfer AddFrozen 保存用户提现信息 到账用户id：%s,币种类型 %s,扣取数量 %s", userid, coinId, amount)
		return x_resp.Fail(-1014, "转账失败", nil), nil
	}

	//减冻结
	if !persistence.ReduceUserFrozenAmount(xmysql, userid, coinId, amount) {
		xmysql.Rollback()
		log.Error("Transfer AddFrozen 添加用户资产失败 到账用户id：%s,币种类型 %s,扣取数量 %s", userid, coinId, amount)
		return x_resp.Fail(-1003, "余额不足", nil), nil
	}
	//存储内部转账信息_出账
	if !(persistence.SaveTransfer(xmysql, userid, typess, coinId, 0-amount, "", types.FROZEN_AMOUNT, orderId+"_1", 0, "", 1, 0) > 0) { //添加提现记录
		xmysql.Rollback()
		log.Error("Transfer AddFrozen 保存用户提现信息 到账用户id：%s,币种类型 %s,扣取数量 %s", userid, coinId, amount)
		return x_resp.Fail(-1014, "转账失败", nil), nil
	}

	return x_resp.Success(0), nil
}

func ReFrozen(userid int64, amount float64, orderId string, typess int64, coinId int64) (*x_resp.XRespContainer, *x_err.XErr) {

	coininfo := persistence.GetCoinByCoinId(mysql.Get(), coinId)

	if coininfo.Sortname == "" {
		return x_resp.Fail(-1056, "币种id传输错误", nil), nil
	}

	xmysql := mysql.Begin()
	defer xmysql.Commit()

	//减冻结
	if !persistence.ReduceUserFrozenAmount(xmysql, userid, coinId, amount) {
		xmysql.Rollback()
		log.Error("Transfer ReFrozen 添加用户资产失败 到账用户id：%s,币种类型 %s,扣取数量 %s", userid, coinId, amount)
		return x_resp.Fail(-1003, "余额不足", nil), nil
	}
	//存储内部转账信息_出账
	if !(persistence.SaveTransfer(xmysql, userid, typess, coinId, 0-amount, "", types.FROZEN_AMOUNT, orderId+"_1", 0, "", 1, 0) > 0) { //添加提现记录
		xmysql.Rollback()
		log.Error("Transfer ReFrozen 保存用户提现信息 到账用户id：%s,币种类型 %s,扣取数量 %s", userid, coinId, amount)
		return x_resp.Fail(-1014, "转账失败", nil), nil
	}

	return x_resp.Success(0), nil
}

func CoinList(userId int64) []types.CoinTypeReturn {
	return persistence.GetCoins(mysql.Get(), userId)

}

func CoinListWallet(userId int64) []types.CoinType {
	return persistence.CoinListWallet(mysql.Get(), userId)

}

func GetUserCoinAmount(userid int64) []map[string]interface{} {
	return persistence.GetUserCoinAmount(mysql.Get(), userid)
}

func GetcoinbyId(coinid int64) types.CoinType {
	return persistence.GetCoinByCoinId(mysql.Get(), coinid)
}

func GetcoinbySortname(coinname string) int64 {
	return persistence.GetCoinIdforName(mysql.Get(), coinname)
}

//提币最大限额
func BigNum(id int64) float64 {
	return persistence.BigNum(mysql.Get(), id)

}

//提币手续费
func Fee(id int64) float64 {
	return persistence.Fee(mysql.Get(), id)
}

//提币手续费
func GetUrl(id int64) string {
	return persistence.Geturl(mysql.Get(), id)
}

func GetCoinInfo(id int64) types.CoinPirce {
	return persistence.GetCoinInfo(mysql.Get(), id)
}
