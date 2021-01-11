package service

import (
	"github.com/EducationEKT/xserver/x_err"
	"github.com/EducationEKT/xserver/x_http/x_resp"
	"github.com/hwhc/hlc_server/log"
	"github.com/hwhc/hlc_server/mysql"
	"github.com/hwhc/hlc_server/persistence"
	"github.com/hwhc/hlc_server/util"
	"github.com/shopspring/decimal"
)

/**
* @Description
* @Author Mikael
* @Date 2021/1/11 16:58
* @Version 1.0
**/


//用户资产增值
func UserAmountIncr(coinId int64)  (*x_resp.XRespContainer, *x_err.XErr)  {

    //今日价格
	nowAmountPrice := persistence.GetPriceByDate(mysql.Get(),coinId,util.Datestr())
	nowAmountPriceDecimal:= decimal.NewFromFloat(nowAmountPrice)
	if nowAmountPriceDecimal.LessThanOrEqual(decimal.NewFromFloat(0.0)){
		log.Error("UserAmountIncr nowAmountPrice 异常  nowAmountPrice : %.8f ",nowAmountPrice)
		return x_resp.Fail(-301,"UserAmountIncr nowAmountPrice 异常",nil), nil
	}

	//昨日价格
	beforeAmountPrice := persistence.GetPriceByDate(mysql.Get(),coinId,util.GetYestdayDateStr())
	beforeAmountPriceDecimal:= decimal.NewFromFloat(beforeAmountPrice)
	if beforeAmountPriceDecimal.LessThanOrEqual(decimal.NewFromFloat(0.0)){
		log.Error("UserAmountIncr beforeAmountPriceDecimal 异常  beforeAmountPrice : %.8f ",beforeAmountPrice)
		return x_resp.Fail(-302,"UserAmountIncr beforeAmountPriceDecimal 异常",nil), nil
	}

	var lastId int64
	var limit int64 = 5000

	for{
		list,err := persistence.GetUserAmountListByLimit(mysql.Get(),coinId,lastId,limit)
		if err != nil{
			log.Error("UserAmountIncr GetUserAmountListByLimit err : %v ",err)
			return x_resp.Fail(-303,"UserAmountIncr GetUserAmountListByLimit err",nil), nil
		}
		size:= len(list)
		if size == 0{
			break
		}

		lastId = list[size-1].Id //当前轮最后一个id

		for _,userAmount := range list {

			amountTotalDecimal := decimal.NewFromFloat(userAmount.Amount) //当前资产
			nowComputePrice:= amountTotalDecimal.Mul(nowAmountPriceDecimal) //今天计算后的
			beforeComputePrice:= amountTotalDecimal.Mul(beforeAmountPriceDecimal) //昨天计算后的
			incrValue,_:= nowComputePrice.Sub(beforeComputePrice).Truncate(5).Float64() //增值计算

			err  = persistence.AddAmountIncrRecord(mysql.Get(),userAmount.UserId,coinId,userAmount.IsShop,userAmount.Amount,incrValue,nowAmountPrice,beforeAmountPrice)
			if err != nil{
				log.Error("UserAmountIncr AddAmountIncrRecord err : %v ",err)
			}
		}
	}

	return x_resp.Success(nil), nil
}