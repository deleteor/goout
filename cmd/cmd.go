package cmd

import (
	"io"
	"os"
	"time"

	"github.com/keep-fool/goout/cmd/client"
	"github.com/keep-fool/goout/cmd/server"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	debug      bool
	configfile string
	logPath    string

	rootCmd = cobra.Command{
		Use:   "goout",
		Short: "",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			logrus.SetFormatter(&logrus.JSONFormatter{})
			if debug {
				logrus.SetLevel(logrus.DebugLevel)
				logrus.Debug("debug mode")
			}
			if logPath != "" {
				setLogger(logPath)
			}
		},
	}
)

func init() {
	rootCmd.Run = func(cmd *cobra.Command, args []string) { rootCmd.Help() }

	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "D", false, "debug level")
	rootCmd.PersistentFlags().StringVarP(&configfile, "config-file", "c", "", "config file path.")
	rootCmd.PersistentFlags().StringVarP(&logPath, "log-path", "l", "", "log file path.")
}

// Execute cmd main
func Execute() {
	rootCmd.AddCommand(server.Cmd())
	rootCmd.AddCommand(client.Cmd())

	rootCmd.Execute()
}

func setLogger(path string) {
	// path := "./bundles/log/go.log"
	/* 日志轮转相关函数
	`WithLinkName` 为最新的日志建立软连接
	`WithRotationTime` 设置日志分割的时间，隔多久分割一次
	WithMaxAge 和 WithRotationCount二者只能设置一个
	 `WithMaxAge` 设置文件清理前的最长保存时间
	 `WithRotationCount` 设置文件清理前最多保存的个数
	*/
	// 下面配置日志每隔 1 分钟轮转一个新文件，保留最近 3 分钟的日志文件，多余的自动清理掉。
	writer, _ := rotatelogs.New(
		path+"log.%Y%m%d",
		// rotatelogs.WithLinkName(path),
		rotatelogs.WithMaxAge(time.Duration(24*30)*time.Hour),
		rotatelogs.WithRotationTime(time.Duration(24)*time.Hour),
	)
	writers := []io.Writer{
		writer,
		os.Stdout}
	//同时写文件和屏幕
	fileAndStdoutWriter := io.MultiWriter(writers...)
	logrus.SetOutput(fileAndStdoutWriter)
	// logrus.SetOutput(writer)
}
