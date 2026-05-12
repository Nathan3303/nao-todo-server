package conf

import (
	"os"

	"github.com/spf13/viper"
)

var Conf *Config

type Config struct {
	Server  *Server    `yaml:"server"`
	MySQL   *MySQL     `yaml:"mysql"`
	Redis   *Redis     `yaml:"redis"`
	Log     *LogConfig `yaml:"log"`
	Uploads *Uploads   `yaml:"uploads"`
}

type Uploads struct {
	UploadDir   string `yaml:"uploadDir"`   // 文件上传根目录
	AvatarDir   string `yaml:"avatarDir"`   // 头像存储子目录
	MaxFileSize int64  `yaml:"maxFileSize"` // 最大文件大小（字节）
	StaticPath  string `yaml:"staticPath"`  // 静态文件访问路径
}

type LogConfig struct {
	Level         string `yaml:"level"`         // 日志级别：debug, info, warn, error, fatal, panic
	FilePath      string `yaml:"filePath"`      // 日志文件路径
	MaxSize       int    `yaml:"maxSize"`       // 单个日志文件最大大小（MB）
	MaxAge        int    `yaml:"maxAge"`        // 日志文件保留天数
	MaxBackups    int    `yaml:"maxBackups"`    // 保留的日志文件副本数量
	Compress      bool   `yaml:"compress"`      // 是否压缩日志文件
	OutputConsole bool   `yaml:"outputConsole"` // 是否同时输出到控制台
}

type Server struct {
	Ip        string `yaml:"ip"`
	Port      string `yaml:"port"`
	Version   string `yaml:"version"`
	JwtSecret string `yaml:"jwtSecret"`
	GoMaxProc int    `yaml:"goMaxProc"`
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
}
