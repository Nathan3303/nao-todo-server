package flags

import "flag"

type AppFlags struct {
	ConfigFile string
	version    bool
}

var Flags = new(AppFlags)

func ParseFlags() {
	flag.StringVar(&Flags.ConfigFile, "f", "app.config.yaml", "应用配置文件")
	flag.BoolVar(&Flags.version, "v", false, "应用版本")
	flag.Parse()
}
