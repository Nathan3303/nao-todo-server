package globals

import (
	"github.com/bwmarrin/snowflake"
	"gorm.io/gorm"
)

type GlobalVars struct {
	DB       *gorm.DB
	SnowNode *snowflake.Node
}

var Vars = &GlobalVars{}
