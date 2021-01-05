package http

import (
	"github.com/EducationEKT/xserver/x_err"
	"github.com/EducationEKT/xserver/x_http/x_req"
	"github.com/EducationEKT/xserver/x_http/x_resp"
	"github.com/EducationEKT/xserver/x_http/x_router"
	"github.com/hwhc/hlc_server/service"
)

func init() {
	x_router.All("/api/user/getExchangeRecordList", loginFilter, getExchangeRecordList) //闪兑记录
	//x_router.All("/api/user/minTotal", loginFilter, minTotal)
	//x_router.All("/api/user/fee", loginFilter, fees)
	x_router.All("/api/user/exchange", loginFilter, addExchangeRecord)
	x_router.Post("/api/user/getExchangeList", loginFilter, getExchangeList) //闪兑
	//x_router.All("/api/user/exchange/exchangeBuyAmount", loginFilter, exchangeBuyAmount) //查看能闪兑
}
func getExchangeRecordList(req *x_req.XReq) (*x_resp.XRespContainer, *x_err.XErr) {
	userId := req.MustGetInt64("user_id")
	code, ret := service.GetExchangeRecordList(userId)
	if code > 0 {
		return x_resp.Success(ret), nil
	}
	return x_resp.Fail(code, "", nil), nil
}

func getExchangeList(req *x_req.XReq) (*x_resp.XRespContainer, *x_err.XErr) {

	return x_resp.Return(service.GetExchangeList(), nil)
}

func minTotal(req *x_req.XReq) (*x_resp.XRespContainer, *x_err.XErr) {
	//userId := req.Context.Sticker["userId"].(int64)
	//coinId := req.MustGetInt64("coinId")
	return x_resp.Return(320, nil)
}

func fees(req *x_req.XReq) (*x_resp.XRespContainer, *x_err.XErr) {
	//userId := req.Context.Sticker["userId"].(int64)
	//coinId := req.MustGetInt64("coinId")
	return x_resp.Return(0.015, nil)
}
func addExchangeRecord(req *x_req.XReq) (*x_resp.XRespContainer, *x_err.XErr) {
	userId := req.MustGetInt64("user_id")
	num := req.MustGetFloat64("num")
	exchange_id := req.MustGetInt64("exchange_id")
	orderId := req.MustGetString("order_id")
	hlcPrice := req.MustGetFloat64("mp_price")
	return service.AddExchangeRecord(exchange_id, num, userId, orderId,hlcPrice)
}

func exchangeBuyAmount(req *x_req.XReq) (*x_resp.XRespContainer, *x_err.XErr) {
	userId := req.Context.Sticker["userId"].(int64)
	exchangeId := req.MustGetInt64("exchangeId")
	amount := service.ExchangeBuyAmount(userId, exchangeId)
	if amount < 0 {
		return x_resp.Fail(-1090, "", amount), nil
	}
	return x_resp.Success(amount), nil
}
