package ssr_demo

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/BaymaxRice/go-ssr/translater"
	"io"
	"io/ioutil"
	"log"
	"net"
)

type addr struct {
	Ip   string `json:"ip"`
	Port string `json:"port"`
}

type Client struct {
	// 数据转换器
	Converter translater.Converter

	// 本地服务地址
	LocalAddr *net.TCPAddr `json:"local_addr"`

	// 服务器地址
	ServerAddr *net.TCPAddr `json:"server_addr"`
}

type Conf struct {
	// 本地服务地址
	LocalAddr addr `json:"local_addr"`

	// 服务器地址
	ServerAddr addr `json:"server_addr"`

	// 加密方式
	Mode string `json:"mode"`

	Password string `json:"password"`
}

const (
	bufSize = 1024
)

func (c *Client) LoadConf(confPath string) error {
	defaultConf := "client.json"
	if confPath != "" {
		defaultConf = confPath
	}

	readConf, err := ioutil.ReadFile(defaultConf)
	if err != nil {
		return fmt.Errorf("配置文件路径错误")
	}
	conf := Conf{}
	err = json.Unmarshal(readConf, &conf)

	if err != nil {
		return fmt.Errorf("配置文件解析失败")
	}

	c.Converter, err = translater.GetNewTranslater(conf.Mode)
	if err != nil {
		return err
	}

	if conf.Password != "" {
		pd, _ := base64.StdEncoding.DecodeString(conf.Password)
		c.Converter.GenNewPW(pd)
	}

	c.LocalAddr, err = net.ResolveTCPAddr("tcp", conf.LocalAddr.Ip+":"+conf.LocalAddr.Port)
	if err != nil {
		return fmt.Errorf("配置local服务配置失败")
	}
	c.ServerAddr, err = net.ResolveTCPAddr("tcp", conf.ServerAddr.Ip+":"+conf.ServerAddr.Port)
	if err != nil {
		return fmt.Errorf("配置server服务配置失败")
	}

	return nil
}

func (c *Client) Run() error {

	listener, err := net.ListenTCP("tcp", c.LocalAddr)
	if err != nil {
		return fmt.Errorf("启动本地监听失败")
	}
	fmt.Printf("ListenTcp: %v success, LocalAddr:%v", c.LocalAddr, c.LocalAddr)
	defer listener.Close()

	// 获取监听数据连接，处理数据
	for {
		localConn, err := listener.AcceptTCP()
		fmt.Printf("AcceptTCP: %v success\n", localConn)
		if err != nil {
			log.Println(err)
			continue
		}
		// localConn被关闭时直接清除所有数据 不管没有发送的数据
		_ = localConn.SetLinger(0)
		go c.handleConn(localConn)
	}
}

func (c *Client) handleConn(con *net.TCPConn) {
	defer con.Close()

	proxyServer, err := net.DialTCP("tcp", nil, c.ServerAddr)
	if err != nil {
		fmt.Println("连接远程服务器失败" + err.Error())
		return
	}
	fmt.Printf("DialTCP: %v success\n", proxyServer)

	defer proxyServer.Close()

	// Conn被关闭时直接清除所有数据 不管没有发送的数据
	_ = proxyServer.SetLinger(0)

	// 进行转发
	// 从 proxyServer 读取数据发送到 localUser
	go func() {
		err := c.DecodeCopy(proxyServer, con)
		if err != nil {
			// 在 copy 的过程中可能会存在网络超时等 error 被 return，只要有一个发生了错误就退出本次工作
			con.Close()
			proxyServer.Close()
		}
	}()
	// 从 localUser 发送数据发送到 proxyServer，这里因为处在翻墙阶段出现网络错误的概率更大
	_ = c.EncodeCopy(proxyServer, con)
}

func (c *Client) DecodeCopy(con *net.TCPConn, dst io.Writer) error {
	buf := make([]byte, bufSize)
	for {
		readCount, errRead := c.DecodeRead(con, buf)
		fmt.Println("client DecodeCopy ", buf, readCount)
		if errRead != nil {
			if errRead != io.EOF {
				return errRead
			} else {
				return nil
			}
		}
		if readCount > 0 {
			writeCount, errWrite := dst.Write(buf[0:readCount])
			if errWrite != nil {
				return errWrite
			}
			if readCount != writeCount {
				return io.ErrShortWrite
			}
		}
	}
}

func (c *Client) DecodeRead(con *net.TCPConn, bs []byte) (n int, err error) {
	n, err = con.Read(bs)
	if err != nil {
		return
	}
	c.Converter.Decrypt(bs[:n])
	return
}

func (c *Client) EncodeWrite(con *net.TCPConn, bs []byte) (int, error) {
	ret := c.Converter.Encrypt(bs)
	return con.Write(ret)
}

// 从src中源源不断的读取原数据加密后写入到dst，直到src中没有数据可以再读取
func (c *Client) EncodeCopy(con *net.TCPConn, dst io.ReadWriteCloser) error {
	buf := make([]byte, bufSize)
	for {
		readCount, errRead := dst.Read(buf)
		fmt.Println("client EncodeCopy ", string(buf), readCount, errRead)
		if errRead != nil {
			if errRead != io.EOF {
				return errRead
			} else {
				return nil
			}
		}
		if readCount > 0 {
			writeCount, errWrite := c.EncodeWrite(con, buf[0:readCount])
			if errWrite != nil {
				return errWrite
			}
			if readCount != writeCount {
				return io.ErrShortWrite
			}
		}
	}
}
