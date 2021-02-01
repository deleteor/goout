package client

import (
	"fmt"
	"log"
	"net"

	"github.com/keep-fool/goout/configs"
	"github.com/spf13/cobra"
)

// Config 客户端配置
type Config struct {
	Local     string `yaml:"local"`
	Server    string `yaml:"server"`
	Password  string `yaml:"password"`
	Encrytype string `yaml:"encrytype"`
	Protocol  string `yaml:"protocol"`
}

// Cmd 高级收益定时同步
func Cmd() *cobra.Command {
	var (
		configPath string
	)
	cmd := &cobra.Command{
		Use:   "client",
		Short: "客户端",
		Run: func(cmd *cobra.Command, args []string) {
			cfg := configs.GetClientConfig()
			run(cfg)
		},
	}

	cmd.Flags().StringVar(&configPath, "config-path", "config/client.yaml", "配置信息文件")
	return cmd
}

// Client 客户端
func run(c Config) {

	// proxy地址
	serverAddr, err := net.ResolveTCPAddr("tcp", c.Server)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("远程服务器: %s:%d", serverAddr.IP, serverAddr.Port)

	listenAddr, err := net.ResolveTCPAddr("tcp", c.Local)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("监听本地端口: %s:%d ", listenAddr.IP, listenAddr.Port)

	// 监听本地端口
	server, err := net.Listen("tcp", c.Local)
	if err != nil {
		fmt.Printf("Listen failed: %v\n", err)
		return
	}

	for {
		client, err := server.Accept()
		if err != nil {
			fmt.Printf("Accept failed: %v", err)
			continue
		}
		go handleProxy(client)
	}
}

func handleProxy(net.Conn) {

}
