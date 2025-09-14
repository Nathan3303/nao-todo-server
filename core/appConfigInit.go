package core

import (
	"os"

	"naotodoserver/flags"

	"gopkg.in/yaml.v2"
)

type AppConfig struct {
	System struct {
		Ip   string `yaml:"ip"`
		Port int    `yaml:"port"`
	} `yaml:"system"`
	Log struct {
		Dir string `yaml:"dir"`
		App string `yaml:"app"`
	} `yaml:"log"`
	DB struct {
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		Using    string `yaml:"using"`
		Debug    bool   `yaml:"debug"`
	} `yaml:"db"`
	Gin struct {
		Ip   string `yaml:"ip"`
		Port string `yaml:"port"`
		Env  string `yaml:"env"`
	} `yaml:"gin"`
}

var Config AppConfig

func ReadConfigurations() {
	byteData, err := os.ReadFile(flags.Flags.ConfigFile)
	if err != nil {
		// fmt.Println("Can't read config file", flags.Flags.ConfigFile)
		panic(err)
	}
	err = yaml.Unmarshal(byteData, &Config)
	if err != nil {
		// fmt.Println("Can't unmarshal byteData")
		panic(err)
	}
	// fmt.Println(Config)
}
