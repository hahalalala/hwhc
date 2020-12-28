package main

import (
	"encoding/json"
	"fmt"
	"github.com/hwhc/hlc_server/task"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/hwhc/hlc_server/conf"
	_ "github.com/hwhc/hlc_server/http"
	"github.com/hwhc/hlc_server/log"
	"github.com/hwhc/hlc_server/mysql"

	"github.com/EducationEKT/xserver/x_http"

	"github.com/go-zhouxun/xlog"
)

func init() {
	http.HandleFunc("/", x_http.Service)
}

func initService() {
	config := initConfig()
	logger := initLog(config.Log)
	log.InitLog(logger)

	mysql.InitDB(config.MysqlConfig)

	task.StartTask()

	//ethClient := ethclient.NewEthClient(config.GethUrl, logger)

	//service.SetETHClient(ethClient)
}

func main() {
	initService()
	fmt.Printf("server listen on :%d \n", conf.GetConfig().Port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", conf.GetConfig().Port), nil)
	if err != nil {
		fmt.Println(err.Error())
	}
}

func getConfigFilePath() string {
	args := os.Args
	confPath := "conf.json"

	if len(args) > 1 {
		confPath = args[1]
	}
	return confPath
}

func initConfig() *conf.Config {
	path := getConfigFilePath()
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}

	data, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}

	var config conf.Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		panic(err)
	}

	conf.SetConfig(config)

	return &config
}

func initLog(logPath string) xlog.XLog {
	fmt.Println("logPath : ",logPath)
	return xlog.NewDailyLog(logPath)
}
