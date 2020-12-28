package http

import (
	"fmt"
	"github.com/EducationEKT/xserver/x_err"
	"github.com/EducationEKT/xserver/x_http/x_req"
	"github.com/EducationEKT/xserver/x_http/x_resp"
	"github.com/EducationEKT/xserver/x_http/x_router"
	"github.com/hwhc/hlc_server/log"
	"github.com/hwhc/hlc_server/service"
)

func init() {
	x_router.All("/admin/api/autoGenAddressPool", AutoGenAddressPool)       //自动创建地址
	x_router.All("/admin/api/createAddress", CreateAddress)                //创建地址
	x_router.All("/admin/api/transferGoBack", loginFilter, transferGoBack) //提现驳回
	x_router.All("/admin/api/transfer_to",loginFilter, transfer_admin) //后台提现
	x_router.All("/admin/api/transferCheck",loginFilter, transferCheck) //提现校验
	x_router.All("/admin/api/take", loginFilter, takeRecording) //提现记录 预览统计	//Inside
	x_router.All("/admin/api/auditXbbTransfer", loginFilter, auditXbbTransfer) //xbb提币审核接口
}

//xbb提币审核接口
func auditXbbTransfer(req *x_req.XReq) (*x_resp.XRespContainer, *x_err.XErr) {
	orderId := req.MustGetString("order_id")
	userId := req.MustGetInt64("user_id")
	txDesc := req.MustGetString("tx_desc")
	status := req.MustGetInt64("status")
	code ,errmsg:= service.TransferXbb(userId,orderId,status,txDesc)
	if code < 0{
		return x_resp.Fail(code, "", errmsg), nil
	}
	return x_resp.Success(code), nil
}

//hoo提币驳回
func transferGoBack(req *x_req.XReq) (*x_resp.XRespContainer, *x_err.XErr) {
	id := req.MustGetInt64("tran_id")
	userId := req.MustGetInt64("user_id")
	txDesc := req.MustGetString("tx_desc")
	code := service.TransferGoBack(txDesc, userId, id)
	if code < 0 {
		return x_resp.Fail(code, "", nil), nil
	}
	return x_resp.Success(code), nil
}

func takeRecording(req *x_req.XReq) (*x_resp.XRespContainer, *x_err.XErr) {
	page := req.MustGetInt64("page")
	size := req.MustGetInt64("size")
	is_shop := req.MustGetInt64("is_shop")

	data, count := service.TakeRecording(page, size, is_shop)
	return x_resp.Success(map[string]interface{}{
		"data":  data,
		"page":  page,
		"count": count,
	}), nil
}


//自动创建地址
func AutoGenAddressPool(req *x_req.XReq) (*x_resp.XRespContainer, *x_err.XErr) {
	return x_resp.Success(service.GenAddressPool()), nil
}

func CreateAddress(req *x_req.XReq) (*x_resp.XRespContainer, *x_err.XErr) {
	coinname := req.MustGetString("coinname")
	num := req.MustGetString("num")
	return x_resp.Success(service.CreateAddress(coinname, num)), nil
}

//校验
func transferCheck(req *x_req.XReq) (*x_resp.XRespContainer, *x_err.XErr)  {

	log.Info("[debug]hoo transferCheck start ")

	transferId := req.MustGetInt64("tran_id")
	cid := req.MustGetInt64("coin_id")
	userId := req.MustGetInt64("user_id")
	adminname := req.MustGetString("admin_name")

	code, msg := service.Transfer_check(userId, transferId, cid, adminname)
	log.Info(fmt.Sprintf("[debug]hoo transferCheck end,code:%d,msg:%s ,transferId:%d,cid:%d,user_id:%d,adminname:%S",code,msg,transferId,cid,userId,adminname))
	if code < 0 {
		return x_resp.Fail(code, msg, msg), nil
	} else {
		return x_resp.Success(""), nil
	}
}

// 提现
func transfer_admin(req *x_req.XReq) (*x_resp.XRespContainer, *x_err.XErr) {

	log.Info("[debug]hoo transfer_admin start ")

	transferId := req.MustGetInt64("tran_id")
	cid := req.MustGetInt64("coin_id")
	user_id := req.MustGetInt64("user_id")
	coinType := service.GetcoinbyId(cid)
	adminname := req.MustGetString("admin_name")

	log.Info(fmt.Sprintf("[debug] transfer_admin param transferId:%d,cid:%d,user_id:%d,adminname:%v",transferId,cid,user_id,adminname))

	var code int
	var msg string

	if coinType.AppWithdrawal != 1 {
		log.Info(fmt.Sprintf("[debug] transfer_admin no support transferId:%d,cid:%d,user_id:%d,adminname:%v",transferId,cid,user_id,adminname))
		return x_resp.Fail(-10, "暂不支持此提币", nil), nil
	}
	code, msg = service.Transfer_app(user_id, transferId, cid, adminname)

	if code < 0 {
		return x_resp.Fail(code, msg, msg), nil
	} else {
		return x_resp.Success(""), nil
	}

}
