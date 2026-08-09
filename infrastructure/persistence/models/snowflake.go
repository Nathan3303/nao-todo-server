package models

import (
	"os"

	"naotodoserver/consts"

	"github.com/bwmarrin/snowflake"
	"github.com/sirupsen/logrus"
)

// SnowNode Snowflake 节点
var SnowNode *snowflake.Node

// InitSnowflake 初始化 Snowflake 节点
// @param machineID 机器 ID，用于区分不同的节点
// @return 返回初始化后的 Snowflake 节点
func InitSnowflake(machineID int64) *snowflake.Node {
	// 设置纪元时间（可选，避免2039问题）
	// 跨端契约常量，见 consts.SnowflakeEpochMS
	snowflake.Epoch = consts.SnowflakeEpochMS

	// 初始化 Snowflake 节点
	var err error
	SnowNode, err = snowflake.NewNode(machineID)
	if err != nil {
		logrus.Fatal(err)
		os.Exit(1)
	}

	// 返回初始化后的 Snowflake 节点
	return SnowNode
}
