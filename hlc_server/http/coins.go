package http

import (
	"fmt"
	"github.com/EducationEKT/xserver/x_err"
	"github.com/EducationEKT/xserver/x_http/x_req"
	"github.com/EducationEKT/xserver/x_http/x_resp"
	"github.com/EducationEKT/xserver/x_http/x_router"
	"github.com/hwhc/hlc_server/log"
	"github.com/hwhc/hlc_server/service"
	"math"
)

func init() {

	x_router.Get("/user/api/getTransfer", loginFilter,getTransfer)                             //根据orderid，类型查询交易记录
	x_router.Get("/user/api/reduceFreeChargeHL",reduceFreeChargeHL)               //扣除未使用的hl
	x_router.Get("/user/api/transfer", loginFilter, transfer)                      //提现 内部转账
	x_router.Get("/user/api/transfer/list", loginFilter,transferList)             //提现/转账记录
	x_router.Get("/user/api/coins/coinList", loginFilter, coinList)                //获取用户币种余额
	x_router.Get("/user/api/coins/getAddress", loginFilter, getAddress)            //获取地址
	x_router.All("/user/api/transfer_out", loginFilter, transfer_wchat)            //提现
	x_router.All("/user/api/transfer_otc", loginFilter, transfer_otc)              //otc提现
	x_router.Get("/user/api/coin/cal", loginFilter, cal)                           //获取地址
	x_router.Get("/user/api/coins/toAmount", loginFilterToBouns, add)              //添加可用资产
	x_router.Get("/user/api/coins/reAmount", loginFilter, reAmount)                //减可用
	x_router.Get("/user/api/coins/reFrozen", loginFilter, reFrozen)                //减冻结
	x_router.Get("/user/api/coins/frozenToAmount", loginFilter, toFrozen)          //添加可用资产减冻结资产
	x_router.Get("/user/api/coins/mergeCoinns", loginFilter, mergeCoinns)          //组合报单
	x_router.Get("/user/api/coins/shopTranfer", loginFilter, shopTranfer)          //商家转账接口
	x_router.Get("/user/api/coins/amountToFrozen", loginFilter, AddFrozenToAmount) //加冻结减可用
	x_router.Get("/user/api/coins/getAddress", loginFilter, getAddress)            //获取地址

	x_router.Get("/user/api/coins/ValidAddress", validAddress)            //校验地址

}


//获取单笔交易信息
func validAddress(req *x_req.XReq) (*x_resp.XRespContainer, *x_err.XErr) {
	return service.ValidAddress()
}

//获取单笔交易信息
func getTransfer(req *x_req.XReq) (*x_resp.XRespContainer, *x_err.XErr) {
	orderId := req.MustGetString("orderId") //orderId
	typ := req.MustGetInt64("type")         //type
	return service.GetTransfer(orderId, typ)
}

//扣除免费给用户未使用的hl
func reduceFreeChargeHL(req *x_req.XReq) (*x_resp.XRespContainer, *x_err.XErr) {
	log.Error("start reduceFreeChargeHL")

	var coindId int64 = 62         //req.MustGetInt64("coinId")	//币种id
	var spaceTime int64 = 2592000 //req.MustGetInt64("spaceTime") //失效时间，间隔s eg：86400是一天
	return service.ReduceFreeRecharge(coindId, spaceTime)
}

// 提现
func transfer_wchat(req *x_req.XReq) (*x_resp.XRespContainer, *x_err.XErr) {

	cid := req.MustGetInt64("coin_id")
	openid := req.MustGetString("address")
	amount := req.MustGetFloat64("amount")
	order_id := req.MustGetString("order_id")
	userId := req.MustGetInt64("user_id")
	memo := req.MustGetString("memo")

	is_shop := req.MustGetInt64("is_shop")

	coinType := service.GetcoinbyId(cid)
	if coinType.AppWithdrawal != 1 {
		return x_resp.Fail(-10, "暂不支持此提币", nil), nil
	}

	if amount < 0 {
		log.Error("提币数量不能小于0 ,userid: %s ,coinid : %s,num:%s", userId, cid, amount)
		return x_resp.Fail(-1, "转账失败，请联系客服", nil), nil
	}

	return service.Transfer_wchat(order_id, cid, openid, amount, userId, memo, is_shop)
}

// otc提现
func transfer_otc(req *x_req.XReq) (*x_resp.XRespContainer, *x_err.XErr) {

	cid := req.MustGetInt64("coin_id")
	//openid := req.MustGetString("address")
	amount := req.MustGetFloat64("amount")
	order_id := req.MustGetString("order_id")
	userId := req.MustGetInt64("user_id")
	memo := req.MustGetString("memo")

	is_shop := req.MustGetInt64("is_shop")

	coinType := service.GetcoinbyId(cid)
	if coinType.AppWithdrawal != 1 {
		return x_resp.Fail(-10, "暂不支持此提币", nil), nil
	}

	if amount < 0 {
		log.Error("提币数量不能小于0 ,userid: %s ,coinid : %s,num:%s", userId, cid, amount)
		return x_resp.Fail(-1, "转账失败，请联系客服", nil), nil
	}

	return service.Transfer_otc(order_id, cid, amount, userId, memo, is_shop)
}

func cal(req *x_req.XReq) (*x_resp.XRespContainer, *x_err.XErr) {
	userid := req.MustGetInt64("user_id")
	coinid := req.MustGetInt64("coin_id")
	orderid := req.MustGetString("order_id")
	t := req.MustGetInt64("types")
	amount := req.MustGetFloat64("amount")

	return service.ReFrozen(userid, amount, orderid, t, coinid)
}

func reFrozen(req *x_req.XReq) (*x_resp.XRespContainer, *x_err.XErr) {
	userid := req.MustGetInt64("user_id")
	coinid := req.MustGetInt64("coin_id")
	orderid := req.MustGetString("order_id")
	t := req.MustGetInt64("types")
	amount := req.MustGetFloat64("amount")

	return service.ReFrozen(userid, amount, orderid, t, coinid)
}

func reAmount(req *x_req.XReq) (*x_resp.XRespContainer, *x_err.XErr) {
	userid := req.MustGetInt64("user_id")
	coinid := req.MustGetInt64("coin_id")
	orderid := req.MustGetString("order_id")
	t := req.MustGetInt64("types")
	amount := req.MustGetFloat64("amount")
	is_shop := req.MustGetInt64("is_shop")

	if amount<0 {
		log.Info(fmt.Sprintf("资产金额异常 userid: %d ,orderid:%s,amount:%.8f",userid,orderid,amount))
		return x_resp.Fail(-1017,"资产金额异常",nil), nil
	}

	return service.ReAmount(userid, amount, orderid, t, coinid, is_shop)
}

func add(req *x_req.XReq) (*x_resp.XRespContainer, *x_err.XErr) {
	userid := req.MustGetInt64("user_id")
	coinid := req.MustGetInt64("coin_id")
	orderid := req.MustGetString("order_id")
	t := req.MustGetInt64("types")
	amount := req.MustGetFloat64("amount")

	is_shop := req.MustGetInt64("is_shop")

	return service.AddAmount(userid, amount, orderid, t, coinid, is_shop)
}

func AddFrozenToAmount(req *x_req.XReq) (*x_resp.XRespContainer, *x_err.XErr) {
	userid := req.MustGetInt64("user_id")
	coinid := req.MustGetInt64("coin_id")
	orderid := req.MustGetString("order_id")
	t := req.MustGetInt64("types")
	amount := req.MustGetFloat64("amount")
	is_shop := req.MustGetInt64("is_shop")

	return service.AddFrozenToAmount(userid, amount, orderid, t, coinid, is_shop)
}

func shopTranfer(req *x_req.XReq) (*x_resp.XRespContainer, *x_err.XErr) {
	userid := req.MustGetInt64("user_id")
	coinid := req.MustGetInt64("coin_id")
	shop_user_id := req.MustGetInt64("shop_user_id")
	orderid := req.MustGetString("order_id")
	t := req.MustGetInt64("types")
	amount := req.MustGetFloat64("amount")
	to_shop_amount := req.MustGetFloat64("to_shop_amount")

	return service.ShopTranfer(userid, amount, orderid, t, coinid, shop_user_id, to_shop_amount)
}

func mergeCoinns(req *x_req.XReq) (*x_resp.XRespContainer, *x_err.XErr) {
	userid := req.MustGetInt64("user_id")
	coinid := req.MustGetInt64("coin_id")
	re_coin_id := req.MustGetInt64("re_coin_id")
	orderid := req.MustGetString("order_id")
	t := req.MustGetInt64("types")
	amount := req.MustGetFloat64("amount")
	re_amount := req.MustGetFloat64("re_amount")

	is_shop := req.MustGetInt64("is_shop")

	return service.MergeCoinns(userid, amount, orderid, t, coinid, is_shop, re_coin_id, re_amount)
}

func toFrozen(req *x_req.XReq) (*x_resp.XRespContainer, *x_err.XErr) {
	userid := req.MustGetInt64("user_id")
	coinid := req.MustGetInt64("coin_id")
	orderid := req.MustGetString("order_id")
	t := req.MustGetInt64("types")
	amount := req.MustGetFloat64("amount")

	is_shop := req.MustGetInt64("is_shop")

	return service.AddAmounToFrozen(userid, amount, orderid, t, coinid, is_shop)
}

func coinList(req *x_req.XReq) (*x_resp.XRespContainer, *x_err.XErr) {
	userId := req.MustGetInt64("user_id")
	return x_resp.Return(service.CoinList(userId), nil)
}

func GetUserCoinAmount(req *x_req.XReq) (*x_resp.XRespContainer, *x_err.XErr) {
	userId := req.MustGetInt64("user_id")
	return x_resp.Return(service.GetUserCoinAmount(userId), nil)
}

//获取币种地址
func getAddress(req *x_req.XReq) (*x_resp.XRespContainer, *x_err.XErr) {

	userId := req.MustGetInt64("user_id")
	coinname := req.MustGetString("coin_name")
	address := service.GetAddressByUserId(coinname, userId)
	if address == "" {
		return x_resp.Fail(-1189, "地址池不足", nil), nil
	}
	return x_resp.Return(address, nil)

}

//内部转账
func transfer(req *x_req.XReq) (*x_resp.XRespContainer, *x_err.XErr) {
	userId := req.MustGetInt64("user_id")
	t := req.MustGetInt64("coin_id")
	re_user_id := req.MustGetInt64("re_user_id") //转账ID
	amount := req.MustGetFloat64("amount")

	if re_user_id == userId {
		return x_resp.Fail(-1021, "用户不能给自己转账", nil), nil
	}
	if amount <= 0 {
		return x_resp.Fail(-1000, "转账余额不能小于0", nil), nil
	}

	coinType := service.GetcoinbyId(t)

	if coinType.InWithdrawal != 1 {
		return x_resp.Fail(-10, "暂不支持此提币", nil), nil
	}

	return service.Transfer(userId, t, amount, re_user_id, coinType.Sortname)
}

func transferList(req *x_req.XReq) (*x_resp.XRespContainer, *x_err.XErr) {
	userId := req.MustGetInt64("user_id")
	coin_id := req.MustGetInt64("coin_id")
	lastId := req.MustGetInt64("lastId")
	size := req.MustGetInt64("size")
	types := req.MustGetInt64("type")

	if lastId == 0 {
		lastId = math.MaxInt64
	}

	data, count := service.GetTransferList(userId, coin_id,  size, types,lastId)
	return x_resp.Success(map[string]interface{}{
		"data":  data,
		"count": count,
	}), nil
}
