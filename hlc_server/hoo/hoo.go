package hoo

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/EducationEKT/EKT/util"
	"github.com/hwhc/hlc_server/conf"
	"github.com/hwhc/hlc_server/log"
	"io"
	"net/url"
	"strings"
)

type Foo struct {
	Code    string   `json:"code"`
	Data    []string `json:"data"`
	Msg     string   `json:"message"`
	Success bool     `json:"success"`
}

type FooTransfer struct {
	Code string `json:"code"`
	Data struct {
		OrderNo string `json:"order_no"`
	} `json:"data"`
	Msg     string `json:"message"`
	Success bool   `json:"success"`
}


type FooOrder struct {
	Code string `json:"code"`
	Msg     string `json:"message"`
	Data struct {
		OuterOrderNo string `json:"outer_order_no"` //txHash
		OrderNo string `json:"order_no"` //hooNo
		TradeType string `json:"trade_type"`
		CoinName string `json:"coin_name"`
		ChainName string `json:"chain_name"`
		TransactionId string `json:"transaction_id"`
		BlockHeight string `json:"block_height"`
		Confirmations string `json:"confirmations"`
		MinConfirmations string `json:"min_confirmations"`
		FromAddress string `json:"from_address"`
		ToAddress string `json:"to_address"`
		Memo string `json:"memo"`
		Amount string `json:"amount"`
		Fee string `json:"fee"`
		Status string `json:"status"` //success
		CreateAt string `json:"create_at"`
		ProcessAt string `json:"process_at"`
	} `json:"data"`

}

type ValidAddressRes struct {
	Code string `json:"code"`
	Msg     string `json:"message"`
	Data struct {
		CoinName string `json:"coin_name"` //txHash
		Address string `json:"address"` //hooNo
	} `json:"data"`

}

const client_id = "VYs4aiKSg8HPDrgKt3bWThaNrk7294"
const pass = "RkVNF18AxqgS9pGadH8czMjfCrpQmFaQx24GP5BrEqmUeyJgMh"
const urls = "https://hoo.com"

func GetHmacCode(s string) string {
	h := hmac.New(sha256.New, []byte(pass))
	io.WriteString(h, s)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func GetHmacCodePass(s string, pass string) string {
	h := hmac.New(sha256.New, []byte(pass))
	io.WriteString(h, s)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func CreateHcAddress(coinName string, num string) []string {

	if conf.GetConfig().Env == "test" {
		return []string{}
	}

	path := "/api/open/vip/v1/address"
	params := url.Values{
		"client_id": []string{client_id},
		"coin_name": []string{coinName},
		"num":       []string{num},
	}

	param := fmt.Sprintf("client_id=%s&coin_name=%s&num=%s", client_id, coinName, num)
	sign := GetHmacCode(param)
	body := fmt.Sprintf(`{"client_id": "%s", "coin_name": "%s", "num": "%s", "sign": "%s"}`, client_id, coinName, num, sign)
	fmt.Println(body)

	params["sign"] = []string{sign}
	fmt.Println(params.Encode())

	resp, err := util.HttpPost(urls+path, []byte(body))
	fmt.Println(urls+path, string(resp), err)

	if err != nil {
		fmt.Println(err)
		return nil

	}
	log.Info("创建地址返回信息：", string(resp))
	var foo Foo
	err = json.Unmarshal(resp, &foo)
	if err != nil || foo.Code != "10000" || len(foo.Data) <= 0 {
		fmt.Println(err)
		log.Error("创建地址失败，", string(resp))
		return nil

	}

	return foo.Data
}

func TransferHCOut(orderId string, amount string, coinname string, toAddress string, contractAddress string, tokenName string, memo string) (string, string) {

	if conf.GetConfig().Env == "test" {
		return "",""
	}

	log.Info("[debug]hoo TransferHCOut start ")

	path := "/api/open/vip/v1/withdraw"

	params := url.Values{
		"amount":           []string{amount}, //提现金额
		"client_id":        []string{client_id},
		"coin_name":        []string{coinname},
		"contract_address": []string{contractAddress}, //代币合约地址，只有转出代币时，才需要填写
		"fee":              []string{""},              //费用
		"memo":             []string{memo},            //备注
		"order_id":         []string{orderId},         //是	订单号，唯一，最长48位长度
		"send_num":         []string{"1"},             //提现序号，提现币种最后一次提现记录的值+1，即为该字段的值（如果没有提现值为1），Hoo会对改值进行验证
		"token_name":       []string{tokenName},
		"to_address":       []string{toAddress}, //提现地址
	}
	pars, _ := url.QueryUnescape(params.Encode())
	sign := GetHmacCode(pars)
	params.Add("sign", sign)

	bodya, _ := url.ParseQuery(params.Encode()) //转map
	body, _ := json.Marshal(bodya)              //转json

	body_str := strings.Replace(string(body), "[", "", -1)
	body_str = strings.Replace(body_str, "]", "", -1)
	fmt.Printf("body_str:%s \n",body_str)
	log.Info(fmt.Sprintf("[debug]hoo TransferHCOut body_str : %s ",body_str))

	url := urls + path
	resp, err := util.HttpPost(url, []byte(body_str))
	fmt.Printf("httpres  url:%s , resp:%s , err : %v \n",url, string(resp), err)
	log.Info(fmt.Sprintf("[debug]httpres  url:%s , resp:%s , err : %v ",url, string(resp), err))
	if err != nil {
		log.Info("sent to address failed, %s", string(resp))
	}
	var foo FooTransfer
	err = json.Unmarshal(resp, &foo)
	fmt.Printf("foo err : %v , foo code : %s ,foo msg : %s \n", err,foo.Code,foo.Msg)
	log.Info(fmt.Sprintf("[debug]foo err : %v , foo code : %s ,foo msg : %s \n", err,foo.Code,foo.Msg))
	if err != nil || foo.Code != "10000" || foo.Msg != "success" {
		log.Error("sent to address failed, %s", string(resp))
		return "", foo.Msg
	}
	log.Info("[debug]hoo success")
	return foo.Data.OrderNo, foo.Msg
}


func GetOrder(orderId string) FooOrder {

	var fooOrder FooOrder

	if conf.GetConfig().Env == "test" {
		return fooOrder
	}

	path := "/api/open/vip/v1/orderdetail"
	params := url.Values{
		"client_id": []string{client_id},
		"order_id": []string{orderId},
	}

	param := fmt.Sprintf("client_id=%s&order_id=%s", client_id, orderId)
	sign := GetHmacCode(param)
	body := fmt.Sprintf(`{"client_id": "%s", "order_id": "%s", "sign": "%s"}`, client_id, orderId, sign)

	params["sign"] = []string{sign}
	resp, err := util.HttpPost(urls+path, []byte(body))
	if err != nil {
		log.Error(fmt.Sprintf("GetOrder util.HttpPost err : %v ,resp : %s ，body：%s", err,string(resp),body))
		return fooOrder
	}

	err = json.Unmarshal(resp, &fooOrder)
	log.Info(fmt.Sprintf("GetOrder json.Unmarshal  err : %v ,resp : %s ,fooOrder:%v ,body:%s", err,string(resp),fooOrder,body))
	if err != nil || fooOrder.Code != "10000" {
		return fooOrder
	}

	return fooOrder
}


func ValidAddress(address string,coinName string) bool {

	if conf.GetConfig().Env == "test" {
		return false
	}

	path := "/api/open/vip/v1/checkaddress"
	params := url.Values{
		"address": []string{address},
		"client_id": []string{client_id},
		"coin_name": []string{coinName},
	}

	param := fmt.Sprintf("address=%s&client_id=%s&coin_name=%s", address ,client_id, coinName)
	sign := GetHmacCode(param)
	body := fmt.Sprintf(`{"address": "%s","client_id": "%s", "coin_name": "%s" , "sign": "%s"}`,address, client_id, coinName, sign)

	fmt.Printf("body : %s \n",body)
	params["sign"] = []string{sign}
	resp, err := util.HttpPost(urls+path, []byte(body))
	if err != nil {
		fmt.Printf("ValidAddress util.HttpPost err : %v ,resp : %s ，body：%s \n", err,string(resp),body)
		log.Error(fmt.Sprintf("ValidAddress util.HttpPost err : %v ,resp : %s ，body：%s", err,string(resp),body))
		return false
	}

	var validAddressRes ValidAddressRes
	err = json.Unmarshal(resp, &validAddressRes)
	if err != nil || validAddressRes.Code != "10000"  || validAddressRes.Data.Address != address{
		fmt.Printf("ValidAddress util.HttpPost err : %v ,resp : %s ，body：%s ,validAddressRes : %v \n", err,string(resp),body,validAddressRes)
		log.Error(fmt.Sprintf("ValidAddress util.HttpPost err : %v ,resp : %s ，body：%s ,validAddressRes : %v ", err,string(resp),body,validAddressRes))
		return false
	}

	return true
}