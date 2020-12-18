package http

import (
	"fmt"
	"github.com/EducationEKT/xserver/x_err"
	"github.com/EducationEKT/xserver/x_http/x_req"
	"github.com/EducationEKT/xserver/x_http/x_resp"
	"github.com/hwhc/hlc_server/conf"
	"github.com/hwhc/hlc_server/hoo"
	"github.com/hwhc/hlc_server/log"
	"net/url"
	"sort"
)

func loginFilter(req *x_req.XReq) (*x_resp.XRespContainer, *x_err.XErr) {

	var pass string
	if conf.GetConfig().Env == "test"{
		pass = "test_hlc_wallet"
	}else{
		pass = "3CuER2e73JBMyB1tjsyhL6zUNxxF1POt"
	}

	m, _ := url.ParseQuery(string(req.Body))
	var list strlist = make([]string, 0)
	for key, _ := range m {
		list = append(list, key)
	}
	sort.Sort(list)
	str := ""
	for _, xxx := range list {
		if xxx == "sign" {
			continue
		}

		str = fmt.Sprintf("%s&%s=%v", str, xxx, m[xxx][0])

	}
	str = str[1:]
	sign := req.MustGetString("sign")
	correctSign := hoo.GetHmacCodePass(str, pass)
	if correctSign != sign {
		log.Error("签名错误： %s = %s", correctSign, m.Get("sign"))
		fmt.Println("签名错误： %s = %s", correctSign, m.Get("sign"))
		return x_resp.Fail(-7, "签名错误", nil), x_err.New(-5, "签名错误")
	}

	return nil, nil
}

func loginFilterToBouns(req *x_req.XReq) (*x_resp.XRespContainer, *x_err.XErr) {

	var pass string
	if conf.GetConfig().Env == "test"{
		pass = "test_hlc_wallet"
	}else{
		pass = "HQcPJHDE6UR49iaEzCjkg9icuwMgWwrp"
	}

	m, _ := url.ParseQuery(string(req.Body))
	var list strlist = make([]string, 0)
	for key, _ := range m {
		list = append(list, key)
	}
	sort.Sort(list)
	str := ""
	for _, xxx := range list {
		if xxx == "sign" {
			continue
		}

		str = fmt.Sprintf("%s&%s=%v", str, xxx, m[xxx][0])

	}
	str = str[1:]
	sign := req.MustGetString("sign")
	correctSign := hoo.GetHmacCodePass(str, pass)
	if correctSign != sign {
		log.Error("签名错误： %s = %s", correctSign, m.Get("sign"))
		return x_resp.Fail(-7, "签名错误", nil), x_err.New(-5, "签名错误")
	}

	return nil, nil
}
