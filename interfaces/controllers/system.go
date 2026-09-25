package controllers

import (
	"strconv"

	"naotodoserver/consts"
	"naotodoserver/interfaces/types"

	"github.com/gin-gonic/gin"
)

// SystemController 系统配置控制器
type SystemController struct{}

// NewSystemController 创建系统配置控制器实例
func NewSystemController() *SystemController {
	return &SystemController{}
}

// GetSystemConfig 获取系统配置控制器
// @code 8000x
// 下发跨端契约常量（如雪花 Epoch），前端启动时请求获取
func (c *SystemController) GetSystemConfig(ctx *gin.Context) {
	Success(ctx, types.ResponseData{
		Code:    80000,
		Message: "获取系统配置成功",
		Data: types.SystemConfigRes{
			SnowflakeEpoch: strconv.FormatInt(consts.SnowflakeEpochMS, 10),
		},
	})
}
