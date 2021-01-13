package types

/**
* @Description 地址库
* @Author Mikael
* @Date 2020/12/15 14:12
* @Version 1.0
**/

type AmountIncrRecord struct {
	Id                int64  `json:"id"`
	UserId            int64 `json:"user_id"`
	CoinId            int64  `json:"coin_id"`
	Amount            float64 `json:"amount"`
	IncrValue         float64 `json:"incr_value"`
	NowAmountPrice    float64 `json:"now_amount_price"`
	BeforeAmountPrice float64 `json:"before_amount_price"`
	IsShop            int64 `json:"is_shop"`
	CreateTime        string `json:"create_time"`
	UpdateTime        string `json:"update_time"`
}
