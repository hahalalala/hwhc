package service

import (
	"fmt"
	"github.com/EducationEKT/xserver/x_err"
	"github.com/EducationEKT/xserver/x_http/x_resp"
	"github.com/hwhc/hlc_server/log"
	"github.com/hwhc/hlc_server/mysql"
	"github.com/hwhc/hlc_server/persistence"
	"github.com/hwhc/hlc_server/types"
	"github.com/shopspring/decimal"
	"sync"
)

/**
* @Description
* @Author Mikael
* @Date 2021/1/11 16:58
* @Version 1.0
**/
var globalCoinId int64
var globalNowAmountPrice float64 //前一天价格
var globalNowAmountPriceDecimal decimal.Decimal
var globalBeforeAmountPrice float64 //后一天价格
var globalBeforeAmountPriceDecimal decimal.Decimal
var globalNowDate string    //前一天时间
var globalBeforeDate string //后一天时间

//用户资产增值
func UserAmountIncr(coinId int64) (*x_resp.XRespContainer, *x_err.XErr) {

	//初始化参数
	err := initUserAmountIncrGlobalVars(coinId)
	if err != nil {
		log.Error(fmt.Sprintf("UserAmountIncr initUserAmountIncrGlobalVars err : %v ", err))
		return x_resp.Fail(-301, "UserAmountIncr err ", nil), nil
	}

	var limit int64 = 5000
	var lastId int64
	var loop int64

	for {
		list, err := persistence.GetUserAmountListByLimit(mysql.Get(), coinId, lastId, limit)
		if err != nil {
			log.Error("UserAmountIncr GetUserAmountListByLimit err : %v ", err)
			return x_resp.Fail(-303, "UserAmountIncr GetUserAmountListByLimit err", nil), nil
		}
		size := len(list)
		if size == 0 {
			break
		}
		loop++
		lastId = list[size-1].Id //当前轮最后一个id
		log.Info(fmt.Sprintf("UserAmountIncr loop:%d ,lastId :%d", loop, lastId))

		//多线程处理
		multiBatchExecStart(list)

		//单线程处理
		//for _,userAmount := range list {
		//	execStart(userAmount)
		//}
	}

	return x_resp.Success(nil), nil
}

//初始化全局参数
func initUserAmountIncrGlobalVars(coinId int64) error {

	globalNowDate = "2021-01-16"//util.GetYestdayDateStr() //前一天时间
	globalBeforeDate = "2021-01-15"//util.Get2dayBefore()  //后一天时间

	globalCoinId = coinId

	//昨日价格
	globalNowAmountPrice = persistence.GetPriceByDate(mysql.Get(), coinId, globalNowDate)
	globalNowAmountPriceDecimal = decimal.NewFromFloat(globalNowAmountPrice)
	if globalNowAmountPriceDecimal.LessThanOrEqual(decimal.NewFromFloat(0.0)) {
		return fmt.Errorf("initUserAmountIncrGlobalVars globalNowAmountPriceDecimal 异常  globalNowAmountPrice : %.8f ", globalNowAmountPrice)
	}

	//前日价格
	globalBeforeAmountPrice = persistence.GetPriceByDate(mysql.Get(), coinId, globalBeforeDate)
	globalBeforeAmountPriceDecimal = decimal.NewFromFloat(globalBeforeAmountPrice)
	if globalBeforeAmountPriceDecimal.LessThanOrEqual(decimal.NewFromFloat(0.0)) {
		return fmt.Errorf("initUserAmountIncrGlobalVars globalBeforeAmountPriceDecimal 异常  globalBeforeAmountPrice : %.8f ", globalBeforeAmountPrice)
	}

	return nil
}

//多线程处理
func multiBatchExecStart(userAmounts []types.UserAmount) {

	TCount := 20 //线程数量
	taskChan := make(chan types.UserAmount)
	var wg sync.WaitGroup //创建一个sync.WaitGroup

	//1)生产任务
	go func() {
		for _, task := range userAmounts {
			taskChan <- task
		}
		close(taskChan)
	}()

	wg.Add(TCount)

	// 2) 消费任务 启动 TCount 个协程执行任务
	for i := 0; i < TCount; i++ {
		go func() {
			defer func() {
				wg.Done()
			}()
			for task := range taskChan {
				autoStatus := task
				execStart(autoStatus)
			}
		}()
	}

	wg.Wait()
}

func execStart(userAmount types.UserAmount) {

	//判断今天是否执行过
	recordId := persistence.GetAmountIncrRecordId(mysql.Get(), userAmount.UserId, globalCoinId, globalNowDate)
	if recordId > 0 {
		return
	}

	//执行1968050.595
	amountTotalDecimal := decimal.NewFromFloat(userAmount.Amount)                       //当前资产
	yesterdayComputePrice := amountTotalDecimal.Mul(globalNowAmountPriceDecimal)        //昨天计算后的
	beforeComputePrice := amountTotalDecimal.Mul(globalBeforeAmountPriceDecimal)        //前天计算后的
	incrValue, _ := yesterdayComputePrice.Sub(beforeComputePrice).Truncate(5).Float64() //增值计算

	err := persistence.AddAmountIncrRecord(mysql.Get(), userAmount.UserId, globalCoinId, userAmount.IsShop, userAmount.Amount, incrValue, globalNowAmountPrice, globalBeforeAmountPrice, globalNowDate, globalBeforeDate)
	if err != nil {
		log.Error("UserAmountIncr AddAmountIncrRecord err : %v ", err)
	}
}
