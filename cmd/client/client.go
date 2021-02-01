package client

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"github.com/keep-fool/goout/auth"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// Cmd 高级收益定时同步
func Cmd() *cobra.Command {
	var (
		configPath string
	)
	cmd := &cobra.Command{
		Use:   "client",
		Short: "客户端",
		Run: func(cmd *cobra.Command, args []string) {
			cfg := getConfig()
			run(cfg)
		},
	}
	cmd.Flags().StringVar(&configPath, "config-path", "config/client.yaml", "配置信息文件")
	return cmd
}

// Client 客户端
func run(c Config) {

	listenAddr, err := net.ResolveTCPAddr("tcp", c.Local)
	if err != nil {
		logrus.Fatal(err)
	}
	logrus.Infof("监听本地端口: %s:%d ", listenAddr.IP, listenAddr.Port)

	// 监听本地端口
	server, err := net.ListenTCP("tcp", listenAddr)
	if err != nil {
		logrus.Fatalf("Listen failed: %v\n", err)
	}

	for {
		client, err := server.Accept()
		if err != nil {
			logrus.Errorf("Accept failed: %v", err)
			continue
		}
		go handle(client, c)
	}
}

func handle(client net.Conn, cfg Config) {
	defer client.Close()

	// 远程连接
	dstServer, err := dialTCP(cfg.Server)
	if err != nil {
		return
	}
	defer dstServer.Close()

	//所有客户服务端的流都加密,
	cypt, err := auth.Create(cfg.Encrytype, cfg.Password)
	if err != nil {
		logrus.Fatal(err)
	}
	// 和远程端建立安全信道
	var wg sync.WaitGroup
	wg.Add(2)

	if cfg.Protocol == "http" {
		/*Socks5 认证
		VER 本次请求的协议版本号，取固定值 0x05（表示socks 5）
		NMETHODS 客户端支持的认证方式数量，可取值 1~255
		METHODS 可用的认证方式列表
		*/
		cypt.EncodeWrite(dstServer, []byte{0x05, 0x01, 0x00})
		resp := make([]byte, 1024)
		n, err := cypt.DecodeRead(dstServer, resp)
		if err != nil {
			logrus.Fatal(err)
		}
		if n == 0 {
			logrus.Fatal("协议错误,服务器返回为空")
			return
		}
		if resp[1] == 0x00 && n == 2 {
			logrus.Print("success")
		} else {
			logrus.Fatal("协议错误，连接失败")
			return
		}

		// SOCKS5协议提供5种认证方式：
		// 0x00 不需要认证
		// 0x01 GSSAPI
		// 0x02 用户名、密码认证
		// 0x03 - 0x7F由IANA分配（保留）
		// 0x80 - 0xFE为私人方法保留

		buff := make([]byte, 1024)
		n, err = client.Read(buff)
		if err != nil {
			logrus.Error(err)
			return
		}
		fmt.Println("-----", n)
		localReq := buff[:n]
		j := 0
		z := 0
		httpreq := []string{}
		for i := 0; i < n; i++ {
			fmt.Println("i-----", i)
			fmt.Println("buff[i]-----", buff[i])
			if buff[i] == 32 {
				httpreq = append(httpreq, string(buff[j:i]))
				j = i + 1
			}
			if buff[i] == 10 {
				z += 1
			}
		}
		fmt.Println("-----", httpreq)
		dstURI, err := url.ParseRequestURI(httpreq[1])
		if err != nil {
			logrus.Error(err)
			return
		}

		var dstAddr string
		var dstPort = "80"
		dstAddrPort := strings.Split(dstURI.Host, ":")
		if len(dstAddrPort) == 1 {
			dstAddr = dstAddrPort[0]
		} else if len(dstAddrPort) == 2 {
			dstAddr = dstAddrPort[0]
			dstPort = dstAddrPort[1]
		} else {
			logrus.Error("URL parse error!")
			return
		}

		/*Socks5
		VER 0x05 Socks5
		CMD 连接方式，0x01=CONNECT, 0x02=BIND, 0x03=UDP ASSOCIATE
		RSV 保留字段
		ATYP 地址类型，0x01=IPv4，0x03=域名，0x04=IPv6
		DST.ADDR 目标地址
		DST.PORT 目标端口，2字节，网络字节序（network octec order）
		*/
		resp = []byte{0x05, 0x01, 0x00, 0x03}

		resp = append(resp, byte(len([]byte(dstAddr))))
		resp = append(resp, []byte(dstAddr)...)
		// 端口
		dstPortBuff := bytes.NewBuffer(make([]byte, 0))
		dstPortInt, err := strconv.ParseUint(dstPort, 10, 16)
		if err != nil {
			logrus.Fatal(err)
		}
		binary.Write(dstPortBuff, binary.BigEndian, dstPortInt)
		dstPortBytes := dstPortBuff.Bytes() // int为8字节
		resp = append(resp, dstPortBytes[len(dstPortBytes)-2:]...)
		n, err = cypt.EncodeWrite(dstServer, resp)
		if err != nil {
			logrus.Error(dstServer.RemoteAddr(), err)
			return
		}
		n, err = cypt.DecodeRead(dstServer, resp)
		if err != nil {
			logrus.Error(dstServer.RemoteAddr(), err)
			return
		}
		var targetResp [10]byte
		copy(targetResp[:10], resp[:n])
		specialResp := [10]byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
		if targetResp != specialResp {
			logrus.Error("协议错误, 第二次协商返回出错")
			return
		}
		log.Print("认证成功")
		// 转发消息
		go func() {
			defer wg.Done()
			cypt.Encrypt(localReq)
			dstServer.Write(localReq)
		}()
		go func() {
			defer wg.Done()
			auth.SecureCopy(dstServer, client, cypt.Decrypt)
		}()
		wg.Wait()
	} else {
		// 本地的内容copy到远程端
		go func() {
			defer wg.Done()
			auth.SecureCopy(client, dstServer, cypt.Encrypt)
		}()

		// 远程得到的内容copy到源地址
		go func() {
			defer wg.Done()
			auth.SecureCopy(dstServer, client, cypt.Decrypt)
		}()
		wg.Wait()
	}
}

func dialTCP(remoteAddr string) (*net.TCPConn, error) {
	// proxy地址
	serverAddr, err := net.ResolveTCPAddr("tcp", remoteAddr)
	if err != nil {
		logrus.Fatal(err)
	}
	logrus.Infof("远程服务器: %s:%d", serverAddr.IP, serverAddr.Port)
	// 远程连接IO
	dstServer, err := net.DialTCP("tcp", nil, serverAddr)
	if err != nil {
		logrus.Errorf("远程服务器地址连接错误 %v", err)
		return nil, err
	}
	return dstServer, nil
}
