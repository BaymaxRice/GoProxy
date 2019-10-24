package ssr_demo

import "net"

type Client struct {
	// 数据转换器
	Converter *Converter

	// 本地服务地址
	LocalAddr *net.TCPAddr

	// 服务器地址
	ServerAddr *net.TCPAddr
}

