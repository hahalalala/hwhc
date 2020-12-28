package service

import (
	"fmt"
	"github.com/hwhc/hlc_server/hoo"
	"github.com/hwhc/hlc_server/log"
	"github.com/hwhc/hlc_server/mysql"
	"github.com/hwhc/hlc_server/persistence"
	"strconv"
)



func TransferXbb(userId int64, orderId string,status int64,txDesc string) (int, string) {

	//审核开关
	auditCoinSwitch := persistence.GetConfig(mysql.Get(),persistence.AuditCoinKey)
	if auditCoinSwitch != "on"{
		return -2001, "系统维护中....."
	}

	// 验证余额
	trasfer := persistence.TransferbyTxhash(mysql.Get(), userId, persistence.IDR,orderId)
	if trasfer.Id == 0{
		return -2002, "该笔交易不存在"
	}

	xmysql := mysql.Begin()

	if status == 1 {
		//成功、修改状态
		if !persistence.UpdateTransferStatus(xmysql, trasfer.Tx_hash, trasfer.UserId) {
			xmysql.Rollback()
			log.Error("TransferXbb 成功 UpdateTransferStatus fail userId :%d ,orderId :%s  status:%d txDesc:%s",userId,orderId,status,txDesc)
			return -2003, "操作失败，刷新后重试"
		}
	}else{
		//失败退款
		if !persistence.UpTransferStatus(xmysql, txDesc, userId, trasfer.Tx_hash, 0, -1) {
			xmysql.Rollback()
			log.Error("TransferXbb 失败退款 UpTransferStatus fail userId :%d ,orderId :%s  status:%d txDesc:%s",userId,orderId,status,txDesc)
			return -2004,"操作失败，刷新后重试"
		}
		if !persistence.AddUserAmount(xmysql, userId, trasfer.CoinId, 0-trasfer.Amount, 0) {
			xmysql.Rollback()
			log.Error("TransferXbb 失败退款 AddUserAmount fail userId :%d ,orderId :%s  status:%d txDesc:%s",userId,orderId,status,txDesc)
			return -2005,"操作失败，刷新后重试"
		}
		if trasfer.IsShop > 0 {
			if !persistence.AddUserAmount(xmysql, userId, persistence.USDT, trasfer.Fee, 0) {
				xmysql.Rollback()
				return -2006,"操作失败，刷新后重试"
			}
		} else {
			if !persistence.AddUserAmount(xmysql, userId, persistence.HLC, trasfer.Fee, 0) {
				xmysql.Rollback()
				return -2007,"操作失败，刷新后重试"
			}
		}
	}

	xmysql.Commit()

	log.Info(fmt.Sprintf("TransferXbb success ,userId : %d ,orderId %s ,statys : %d ,txDesc :%s ",userId,orderId,status,txDesc))
	return 0, ""
}


func Transfer_app(userId int64, transferId int64, cid int64, adminid string) (int, string) {

	//审核开关
	auditCoinSwitch := persistence.GetConfig(mysql.Get(),persistence.AuditCoinKey)
	if auditCoinSwitch != "on"{
		return -2012, "系统维护中....."
	}

	xmysql := mysql.Begin()
	defer xmysql.Commit()
	// 验证余额    验证余额    验证余额
	trasfer := persistence.TransferbyId(mysql.Get(), userId, transferId, cid)
	log.Info(fmt.Sprintf("[debug] Transfer_app trasfer->trasferId：%d,transferId:%d,cid:%d,adminid : %v", trasfer.Id, transferId, cid,adminid))
	if !persistence.UpdateTransferStatus(xmysql, trasfer.Tx_hash, trasfer.UserId) {
		return -2011, "更新订单状态失败，刷新后重试"
	}

	coinInfo := GetcoinbyId(cid)
	amount := strconv.FormatFloat(0-trasfer.Amount, 'f', 6, 64)
	hooOrderNo, msg := hoo.TransferHCOut(trasfer.Tx_hash, amount, coinInfo.Coinname, trasfer.Address, coinInfo.ContractCddress, coinInfo.Tokenname, trasfer.Memo)

	persistence.UpdateTransferHooOrderNo(xmysql, trasfer.Id, trasfer.UserId, hooOrderNo)
	log.Info(fmt.Sprintf("[debug] 管理员：%v,处理了订单：%d ,hooOrderNo：%s", adminid, transferId, hooOrderNo))

	if hooOrderNo == "" {
		xmysql.Rollback()
		log.Info(fmt.Sprintf("[debug]transfer rollback,hooOrderNo:%s",hooOrderNo))
		return -10010, msg
	} else {
		log.Info("[debug]transfer success")
		return 0, ""
	}
}


func Transfer_check(userId int64, transferId int64, cid int64, adminid string) (int, string) {

	trasfer := persistence.TransferbyId(mysql.Get(), userId, transferId, cid)
	if trasfer.Id == 0{
		return -10012,"该笔交易不存在，刷新后重试"
	}
	if trasfer.Tx_status == -1{
		log.Info(fmt.Sprintf("[debug]Transfer_check status is -1 transferId:%d ,cid :%d,adminid:%s",transferId,cid,adminid))
		return -10012,"该笔交易已被驳回，刷新后重试"
	}
	if trasfer.Tx_status == 1{
		log.Info(fmt.Sprintf("[debug]Transfer_check status is 1 transferId:%d ,cid :%d,adminid:%s",transferId,cid,adminid))
		return 0,""
	}

	//查询hoo钱包状态
	fooOrder := hoo.GetOrder(trasfer.Tx_hash)
	if fooOrder.Data.OuterOrderNo == "" || fooOrder.Data.OrderNo == "" {
		return -10012,"该笔交易在hoo未查询到"
	}

	hooOrderNo := fooOrder.Data.OrderNo

	if fooOrder.Data.Status == "success"{
		xmysql := mysql.Begin()

		//更新状态
		if !persistence.UpdateTransferStatus(xmysql, trasfer.Tx_hash, trasfer.UserId) {
			xmysql.Rollback()
			log.Info(fmt.Sprintf("[debug] 管理员：%s,Transfer_check 更新订单状态失败 transferId ：%d ,hooOrderNo：%s ,userid:%d , cid:%d ", adminid, transferId, hooOrderNo,userId,cid))
			return -2011, "更新订单状态失败，刷新后重试"
		}
		//更新hoo地址
		if !persistence.UpdateTransferHooOrderNo(xmysql, trasfer.Id, trasfer.UserId, hooOrderNo){
			xmysql.Rollback()
			log.Info(fmt.Sprintf("[debug] 管理员：%s,Transfer_check 更新hooOrderNo失败 transferId ：%d ,hooOrderNo：%s ,userid:%d , cid:%d ", adminid, transferId, hooOrderNo,userId,cid))
			return -10010, "更新hooOrderNo失败，刷新后重试"
		}

		log.Info("[debug]Transfer_check success")
		xmysql.Commit()
		return 0, ""
	}else{
		return -10012,"该笔交易在hoo交易未完成，刷新后重试"
	}

}
