package types

/**
* @Description 地址库
* @Author Mikael
* @Date 2020/12/15 14:12
* @Version 1.0
**/

type AddressPool struct {
	Id       int64  `json:"id"`
	Address  string `json:"address"`
	Status   int64  `json:"status"`
	Coinname string `json:"coinname"`
	UserId   string `json:"user_id"`
}
