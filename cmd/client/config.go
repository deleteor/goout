package client

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
)

// Config 客户端配置
type Config struct {
	Local     string `yaml:"local"`
	Server    string `yaml:"server"`
	Password  string `yaml:"password"`
	Encrytype string `yaml:"encrytype"`
	Protocol  string `yaml:"protocol"`
}

// getConfig 初始化 加载系统配置
func getConfig() Config {
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
	c := Config{}
	if err := yaml.Unmarshal(yamlFile, &c); err != nil {
		panic(err)
	}
	return c
}
