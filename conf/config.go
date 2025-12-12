package conf

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

var Conf *Config

type Config struct {
	Server *Server `yaml:"server"`
	MySQL  *MySQL  `yaml:"mysql"`
	Redis  *Redis  `yaml:"redis"`
}

type Server struct {
	Ip        string `yaml:"ip"`
	Port      string `yaml:"port"`
	Version   string `yaml:"version"`
	JwtSecret string `yaml:"jwtSecret"`
}

type MySQL struct {
	Host      string `yaml:"host"`
	Port      string `yaml:"port"`
	Database  string `yaml:"database"`
	Username  string `yaml:"username"`
	Password  string `yaml:"password"`
	Charset   string `yaml:"charset"`
	ParseTime string `yaml:"parseTime"`
	Loc       string `yaml:"loc"`
}

type Redis struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	DB       int    `yaml:"database"`
	Password string `yaml:"password"`
	Network  string `yaml:"network"`
}

func InitConfig() {
	// @step 1. 获取当前目录
	configDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	// @step 2. 设置配置文件名称预类型，并拼接配置文件路径
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(configDir + "/conf")

	// @step 3. 读取配置文件
	err = viper.ReadInConfig()
	if err != nil {
		panic(err)
	}

	// @step 4. 将配置文件内容解析到 Conf 中
	err = viper.Unmarshal(&Conf)
	if err != nil {
		panic(err)
	}
	fmt.Println(Conf.Redis)
}
