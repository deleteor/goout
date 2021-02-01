package configs

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/keep-fool/goout/cmd/client"
	"github.com/keep-fool/goout/cmd/server"
	yaml "gopkg.in/yaml.v2"
)

type global struct {
}

// GetClientConfig 初始化 加载系统配置
func GetClientConfig() client.Config {
	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	adsPath := strings.Replace(pwd, "\\", "/", -1)
	f := filepath.Join(adsPath+"/configs/", "client.yaml")
	yamlFile, err := ioutil.ReadFile(f)
	if err != nil {
		panic(err)
	}
	c := client.Config{}
	if err := yaml.Unmarshal(yamlFile, &c); err != nil {
		panic(err)
	}
	return c
}

// GetServerConfig 初始化 加载系统配置
func GetServerConfig() server.Config {
	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	adsPath := strings.Replace(pwd, "\\", "/", -1)
	f := filepath.Join(adsPath+"/configs/", "server.yaml")
	yamlFile, err := ioutil.ReadFile(f)
	if err != nil {
		panic(err)
	}
	c := server.Config{}
	if err := yaml.Unmarshal(yamlFile, &c); err != nil {
		panic(err)
	}
	return c
}
