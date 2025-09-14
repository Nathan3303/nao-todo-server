package flags

import (
	"flag"
)

type AppFlags struct {
	ConfigFile    string
	Version       bool
	AutoMigration bool
}

var Flags = new(AppFlags)

func ParseFlags() {
	flag.StringVar(&Flags.ConfigFile, "f", "app.config.yaml", "应用配置文件")
	flag.BoolVar(&Flags.AutoMigration, "m", false, "数据库自动迁移")
	flag.BoolVar(&Flags.Version, "v", false, "应用版本")
	flag.Parse()
}
