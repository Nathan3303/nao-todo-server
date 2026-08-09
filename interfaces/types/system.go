package types

// SystemConfigRes 系统配置响应
type SystemConfigRes struct {
	// SnowflakeEpoch 雪花 ID 纪元（Epoch），单位毫秒
	// 字符串类型：与全项目 ID 字符串化约定一致，避免 JS 大整数精度问题
	SnowflakeEpoch string `json:"snowflakeEpoch"`
}
