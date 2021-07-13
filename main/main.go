package main

import (
	"flag"
	"fmt"
	go_ssr "github.com/BaymaxRice/go-ssr"
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
	fmt.Println("usage:   go-ssr  [-q] [-conf=<filepath>] [-s] [-b]")
	flag.PrintDefaults()
}

func main() {
	if cmdArgs.isServer {
		server := go_ssr.Server{}
		err := server.LoadConf(cmdArgs.conf)
		if err != nil {
			fmt.Println(err)
			return
		}
		_ = server.Run()
	}

	// 客户端程序
	client := go_ssr.Client{}
	err := client.LoadConf(cmdArgs.conf)
	if err != nil {
		fmt.Println(err)
		return
	}

	_ = client.Run()
}
