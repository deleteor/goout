package server

import (
	"log"
	"net"
	"sync"

	"github.com/keep-fool/goout/auth"
	"github.com/keep-fool/goout/socks5"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// Cmd 高级收益定时同步
func Cmd() *cobra.Command {
	var (
		configPath string
	)
	cmd := &cobra.Command{
		Use:   "server",
		Short: "服务端",
		Run: func(cmd *cobra.Command, args []string) {
			cfg := getConfig()
			run(cfg)
		},
	}

	cmd.Flags().StringVar(&configPath, "config-path", "config/server.yaml", "配置文件地址")
	return cmd
}

// run 服务端
func run(c Config) {

	// 监听客户端
	listenAddr, err := net.ResolveTCPAddr("tcp", c.Local)
	if err != nil {
		logrus.Fatal(err)
	}

	listener, err := net.ListenTCP("tcp", listenAddr)
	if err != nil {
		logrus.Fatal(err)
	}
	defer listener.Close()

	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			logrus.Fatal(err)
		}
		go handle(conn, c)
	}
}

func handle(client *net.TCPConn, c Config) {
	defer client.Close()

	//所有客户服务端的流都加密,
	cypt, err := auth.Create(c.Encrytype, c.Password)
	if err != nil {
		log.Fatal(err)
	}

	// 初始化一个字符串buff
	buff := make([]byte, 255)

	// 认证协商
	var proto socks5.ProtocolVersion
	n, err := cypt.DecodeRead(client, buff) //解密
	resp, err := proto.HandleHandshake(buff[0:n])
	cypt.EncodeWrite(client, resp) //加密
	if err != nil {
		log.Print(client.RemoteAddr(), err)
		return
	}

	//获取客户端代理的请求
	var request socks5.Socks5Resolution
	n, err = cypt.DecodeRead(client, buff)
	resp, err = request.LSTRequest(buff[0:n])
	cypt.EncodeWrite(client, resp)
	if err != nil {
		log.Print(client.RemoteAddr(), err)
		return
	}

	log.Println(client.RemoteAddr(), request.DSTDOMAIN, request.DSTADDR, request.DSTPORT)

	// 连接真正的远程服务
	dstServer, err := net.DialTCP("tcp", nil, request.RAWADDR)
	if err != nil {
		log.Print(client.RemoteAddr(), err)
		return
	}
	defer dstServer.Close()

	wg := new(sync.WaitGroup)
	wg.Add(2)

	// 本地的内容copy到远程端
	go func() {
		defer wg.Done()
		auth.SecureCopy(client, dstServer, cypt.Decrypt)
	}()

	// 远程得到的内容copy到源地址
	go func() {
		defer wg.Done()
		auth.SecureCopy(dstServer, client, cypt.Encrypt)
	}()
	wg.Wait()
}
