package models

import (
	"os"
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/sirupsen/logrus"
)

var SnowNode *snowflake.Node

func InitSnowflake(machineID int64) *snowflake.Node {
	// 设置纪元时间（可选，避免2039问题）
	snowflake.Epoch = time.Date(2023, 3, 1, 0, 0, 0, 0, time.UTC).Unix() * 1000

	var err error
	SnowNode, err := snowflake.NewNode(machineID)
	if err != nil {
		logrus.Fatal(err)
		os.Exit(1)
	}
	return SnowNode
}
