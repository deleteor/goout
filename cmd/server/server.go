package server

import (
	"github.com/keep-fool/goout/configs"
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
			cfg := configs.GetServerConfig()
			Run(cfg)
		},
	}

	cmd.Flags().StringVar(&configPath, "config-path", "config/server.yaml", "配置文件地址")
	return cmd
}

// Config 客户端配置
type Config struct {
	Local     string `yaml:"local"`
	Server    string `yaml:"server"`
	Password  string `yaml:"password"`
	Encrytype string `yaml:"encrytype"`
	Protocol  string `yaml:"protocol"`
}

// Run 服务端
func Run(c Config) {

}
