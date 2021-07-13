package go_ssr

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/BaymaxRice/go-ssr/convertor"
	"github.com/BurntSushi/toml"
	"io"
	"io/ioutil"
	"log"
	"net"
)

type Server struct {
	// 数据转换器
	Converter convertor.Convertor

	// 本地服务地址
	LocalAddr *net.TCPAddr
}

type ServerConvertor struct {
	Mode     string `toml:"mode"`
	Password string `toml:"password"`
}

type ServerConf struct {
	// 本地服务地址
	LocalAddr addr            `toml:"local_addr"`
	Convertor ServerConvertor `toml:"convertor"`
}

func (s *Server) LoadConf(confPath string) error {
	defaultConf := "server.toml"
	if confPath != "" {
		defaultConf = confPath
	}

	conf := &ServerConf{}
	_, err := toml.DecodeFile(defaultConf, conf)
	if err != nil {
		return err
	}

	s.Converter, err = convertor.GetNewConvertor(conf.Convertor.Mode)
	if err != nil {
		return err
	}

	// 如果配置文件已经有密码，则根据此密码生成加密解密秘钥
	if conf.Convertor.Password != "" {
		pd, _ := base64.StdEncoding.DecodeString(conf.Convertor.Password)
		s.Converter.GenNewPW(pd)
	} else {
		conf.Convertor.Password = base64.StdEncoding.EncodeToString(s.Converter.GetPW())
		data, err := json.Marshal(conf)
		if err != nil {
			return fmt.Errorf("序列化配置失败")
		}
		_ = ioutil.WriteFile(defaultConf, data, 755)
	}

	s.LocalAddr, err = net.ResolveTCPAddr("tcp", conf.LocalAddr.Ip+":"+conf.LocalAddr.Port)
	if err != nil {
		return fmt.Errorf("配置local服务配置失败")
	}

	return nil
}

func (s *Server) Run() error {
	listener, err := net.ListenTCP("tcp", s.LocalAddr)
	if err != nil {
		return fmt.Errorf("启动本地监听失败")
	}
	fmt.Printf("ListenTcp: %v success, LocalAddr:%v\n", s.LocalAddr, s.LocalAddr)
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
		go s.handleConn(localConn)
	}
}

func (s *Server) handleConn(con *net.TCPConn) {
	defer con.Close()

	buf := make([]byte, 256)

	// 第一个字段VER代表Socks的版本，Socks5默认为0x05，其固定长度为1个字节
	_, err := s.DecodeRead(con, buf)
	// 只支持版本5
	if err != nil || buf[0] != 0x05 {
		fmt.Println("数据协议版本不对")
		return
	}

	_, _ = s.EncodeWrite(con, []byte{0x05, 0x00})

	// 获取真正的远程服务的地址
	n, err := s.DecodeRead(con, buf)
	// n 最短的长度为7 情况为 ATYP=3 DST.ADDR占用1字节 值为0x0
	fmt.Printf("获取真正服务器：n：%d, err: %+v \n", n, err)
	if err != nil || n < 7 {
		fmt.Println("获取真正服务器出错")
		return
	}

	// CMD代表客户端请求的类型，值长度也是1个字节，有三种类型
	// CONNECT X'01'
	if buf[1] != 0x01 {
		// 目前只支持 CONNECT
		fmt.Println("客户端请求类型出错")
		return
	}

	var dIP []byte
	// aType 代表请求的远程服务器地址类型，值长度1个字节，有三种类型
	switch buf[3] {
	case 0x01:
		//	IP V4 address: X'01'
		dIP = buf[4 : 4+net.IPv4len]
	case 0x03:
		//	DOMAINNAME: X'03'
		ipAddr, err := net.ResolveIPAddr("ip", string(buf[5:n-2]))
		fmt.Printf("ip:%+v, err: %+v\n", ipAddr, err)
		if err != nil {
			return
		}
		dIP = ipAddr.IP
	case 0x04:
		//	IP V6 address: X'04'
		dIP = buf[4 : 4+net.IPv6len]
	default:
		return
	}
	dPort := buf[n-2:]
	fmt.Println("port:", dPort)
	dstAddr := &net.TCPAddr{
		IP:   dIP,
		Port: int(binary.BigEndian.Uint16(dPort)),
	}

	// 连接真正的远程服务
	dstServer, err := net.DialTCP("tcp", nil, dstAddr)
	if err != nil {
		fmt.Printf("连接真正服务器出错, err:%+v\n", err)
		return
	} else {
		fmt.Println("连接真正服务器成功")
		defer dstServer.Close()
		// Conn被关闭时直接清除所有数据 不管没有发送的数据
		dstServer.SetLinger(0)

		// 响应客户端连接成功
		/**
		  +----+-----+-------+------+----------+----------+
		  |VER | REP |  RSV  | ATYP | BND.ADDR | BND.PORT |
		  +----+-----+-------+------+----------+----------+
		  | 1  |  1  | X'00' |  1   | Variable |    2     |
		  +----+-----+-------+------+----------+----------+
		*/
		// 响应客户端连接成功
		_, _ = s.EncodeWrite(con, []byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
	}

	// 进行转发
	// 从 localUser 读取数据发送到 dstServer
	go func() {
		err := s.DecodeCopy(con, dstServer)
		if err != nil {
			fmt.Printf("请求远程服务出错，err: %+v\n", err)
			// 在 copy 的过程中可能会存在网络超时等 error 被 return，只要有一个发生了错误就退出本次工作
			con.Close()
			dstServer.Close()
		}
	}()
	// 从 dstServer 读取数据发送到 localUser，这里因为处在翻墙阶段出现网络错误的概率更大
	_ = s.EncodeCopy(dstServer, con)
}

func (s *Server) DecodeCopy(con *net.TCPConn, dst io.Writer) error {
	buf := make([]byte, bufSize)
	for {
		readCount, errRead := s.DecodeRead(con, buf)
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

func (s *Server) DecodeRead(con *net.TCPConn, bs []byte) (n int, err error) {
	n, err = con.Read(bs)
	fmt.Println("read data:", bs)
	if err != nil {
		return
	}
	s.Converter.Decrypt(bs[:n])
	return
}

func (s *Server) EncodeWrite(con *net.TCPConn, bs []byte) (int, error) {
	ret := s.Converter.Encrypt(bs)
	return con.Write(ret)
}

// 从src中源源不断的读取原数据加密后写入到dst，直到src中没有数据可以再读取
func (s *Server) EncodeCopy(con *net.TCPConn, dst io.ReadWriteCloser) error {
	buf := make([]byte, bufSize)
	for {
		readCount, errRead := dst.Read(buf)
		if errRead != nil {
			if errRead != io.EOF {
				return errRead
			} else {
				return nil
			}
		}
		if readCount > 0 {
			writeCount, errWrite := s.EncodeWrite(con, buf[0:readCount])
			if errWrite != nil {
				return errWrite
			}
			if readCount != writeCount {
				return io.ErrShortWrite
			}
		}
	}
}
