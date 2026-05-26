package conf

import (
	"os"
	"strconv"

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
	configDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(configDir + "/conf")

	err = viper.ReadInConfig()
	if err != nil {
		panic(err)
	}

	err = viper.Unmarshal(&Conf)
	if err != nil {
		panic(err)
	}

	overrideWithEnv()
}

func overrideWithEnv() {
	if env := os.Getenv("MYSQL_HOST"); env != "" {
		Conf.MySQL.Host = env
	}
	if env := os.Getenv("MYSQL_PORT"); env != "" {
		Conf.MySQL.Port = env
	}
	if env := os.Getenv("MYSQL_USER"); env != "" {
		Conf.MySQL.Username = env
	}
	if env := os.Getenv("MYSQL_PASSWORD"); env != "" {
		Conf.MySQL.Password = env
	}
	if env := os.Getenv("MYSQL_DATABASE"); env != "" {
		Conf.MySQL.Database = env
	}

	if env := os.Getenv("REDIS_HOST"); env != "" {
		Conf.Redis.Host = env
	}
	if env := os.Getenv("REDIS_PORT"); env != "" {
		Conf.Redis.Port = env
	}
	if env := os.Getenv("REDIS_PASSWORD"); env != "" {
		Conf.Redis.Password = env
	}
	if env := os.Getenv("REDIS_DB"); env != "" {
		if db, err := strconv.Atoi(env); err == nil {
			Conf.Redis.DB = db
		}
	}

	if env := os.Getenv("JWT_SECRET"); env != "" {
		Conf.Server.JwtSecret = env
	}
}
