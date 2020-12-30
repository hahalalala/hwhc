package price

import (
	"github.com/hwhc/hlc_server/mysql"
	"github.com/hwhc/hlc_server/persistence"
	"github.com/hwhc/hlc_server/util"
	"time"

	"github.com/go-zhouxun/xutil/xtime"
)

const INTERVAL = 15 * time.Second

func Start() {
	go recordUSDT()
}

func recordUSDT() {
	for {
		dateStr := util.Datestr()
		t := xtime.Now()
		hlcPrice ,idrPrice ,vndPrice := persistence.GetHLCPrice()
		persistence.SavePrice(mysql.Get(), persistence.USDT, 1, t, dateStr)
		persistence.SavePrice(mysql.Get(), persistence.HLC, hlcPrice, t, dateStr)
		persistence.SavePrice(mysql.Get(), persistence.IDR, idrPrice, t, dateStr)
		persistence.SavePrice(mysql.Get(), persistence.VND, vndPrice, t, dateStr)

		time.Sleep(INTERVAL)

	}
}
