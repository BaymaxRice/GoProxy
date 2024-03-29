package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/BaymaxRice/GoProxy"
)

type CmdArgs struct {
	isServer    bool
	backRunning bool
	conf        string
}

var cmdArgs CmdArgs

func init() {
	flag.BoolVar(&cmdArgs.isServer, "s", false, "输入-s：启动服务端程序")
	// flag.BoolVar(&cmdArgs.isClient, "c", true, "输入-c：启动客户端程序")
	flag.BoolVar(&cmdArgs.backRunning, "b", false, "输入-b：启动后台运行")
	flag.StringVar(&cmdArgs.conf, "conf", "", "设置客户端或者服务端配置文件(默认取当前目录下的client.json)")

	flag.Usage = usage
	flag.Parse()
}

func usage() {
	log.Println("usage:   GoProxy [-conf=<filepath>] [-s] [-b]")
	flag.PrintDefaults()
}

func main() {
	if cmdArgs.isServer {
		server := GoProxy.Server{}
		err := server.LoadConf(cmdArgs.conf)
		if err != nil {
			panic(fmt.Sprintf("load server conf failed, err: %+v", err))
		}
		_ = server.Run()
	}

	// 客户端程序
	client := GoProxy.Client{}
	err := client.LoadConf(cmdArgs.conf)
	if err != nil {
		panic(fmt.Sprintf("load client conf failed, err: %+v", err))
	}

	_ = client.Run()
}
