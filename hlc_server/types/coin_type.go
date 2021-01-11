package types

type CoinType struct {
	Id              int64   `json:"id"`
	Url             string  `json:"url"`
	Coinname        string  `json:"coinname"`
	ContractCddress string  `json:"contract_address"`
	Sortname        string  `json:"sortname"`
	TransferFee     int64   `json:"transfer_fee"`
	Status          int64   `json:"status"`
	Tokenname       string  `json:"tokenname"`
	Type            string  `json:"type"`
	ChainNum        int64   `json:"chain_num"`
	UserId          string  `json:"userId"`
	Amount          float64 `json:"amount"`
	FrozenAmount    float64 `json:"frozen_amount"`
	TzAmount        float64 `json:"tz_amount"`
	CoinQCPrice     float64 `json:"coinQCPrice"`
	Recharge        int64   `json:"recharge"`
	Withdrawal      int64   `json:"withdrawal"`
	InWithdrawal    int64   `json:"in_withdrawal"`
	HzAmount        float64 `json:"hz_amount"`
	TransferBigNum  string  `json:"transfer_big_num"`
	MarketPath      string  `json:"market_path"`
	AppRecharge     int64   `json:"app_recharge"`
	AppWithdrawal   int64   `json:"app_withdrawal"`
	IsTransfer      int64   `json:"is_transfer"`
}

type CoinTypeReturn struct {
	Id               int64   `json:"id"`
	Url              string  `json:"url"`
	Coinname         string  `json:"coinname"`
	Sortname         string  `json:"sortname"`
	Status           int64   `json:"status"`
	Tokenname        string  `json:"tokenname"`
	UserId           string  `json:"userId"`
	Amount           float64 `json:"amount"`
	FrozenAmount     float64 `json:"frozen_amount"`
	CoinQCPrice      float64 `json:"coinQCPrice"`
	InWithdrawal     int64   `json:"in_withdrawal"`
	TransferBigNum   string  `json:"transfer_big_num"`
	MarketPath       string  `json:"market_path"`
	AppRecharge      int64   `json:"app_recharge"`
	AppWithdrawal    int64   `json:"app_withdrawal"`
	IsTransfer       int64   `json:"is_transfer"`
	FrozenUsdtPrice  float64 `json:"frozen_usdt_price"`
	TransferSmallNum int64   `json:"transfer_small_num"`
	TransferFee      float64 `json:"transfer_fee"`
}

type UserAmount struct {
	Id           int64   `json:"id"`
	UserId       int64   `json:"user_id"`
	Type         int64   `json:"type"`
	Amount       float64 `json:"amount"`
	CreateTime   string  `json:"create_time"`
	UpdateTime   string  `json:"update_time"`
	Status       int64   `json:"status"`
	FrozenAmount float64 `json:"frozen_amount"`
	TzAmount     float64 `json:"tz_amount"`
	HzAmount     float64 `json:"hz_amount"`
	IsShop 		 int64   `json:"is_shop"`
}
