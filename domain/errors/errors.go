package errors

import "errors"

// 通用领域错误
var (
	ErrInvalidUserID     = errors.New("用户 ID 无效")
	ErrInvalidID         = errors.New("ID 格式错误")
	ErrInvalidProjectID  = errors.New("清单 ID 格式错误")
	ErrInvalidTaskID     = errors.New("待办任务 ID 无效")
	ErrInvalidTagID      = errors.New("标签 ID 格式错误")
	ErrInvalidCommentID  = errors.New("评论 ID 无效")
	ErrInvalidItemID     = errors.New("检查事项 ID 格式错误")
	ErrInvalidPomodoroID = errors.New("常用番茄工作 ID 无效")
	ErrPasswordMismatch  = errors.New("密码错误")
	ErrEmailExists       = errors.New("该邮箱已被注册")
	ErrUserDeactivated   = errors.New("用户已注销")
	ErrUserInCooldown    = errors.New("操作过于频繁，请30天后再试")
	ErrProjectNotFound   = errors.New("清单不存在")
	ErrUserNotFound      = errors.New("用户不存在")
	ErrTokenExpired      = errors.New("凭证无效")
	ErrSessionNotFound   = errors.New("会话不存在")
	ErrCreateFailed      = errors.New("创建失败")
	ErrDeleteFailed      = errors.New("删除失败")
	ErrIDConflict        = errors.New("ID 冲突：本地生成 ID 与远程已存在实体碰撞，请重新分配")
	// ErrSortIdOverflow 组内排序值已耗尽（组内 max+1 超出 uint16 上界 65535）：
	// 不写入任何数据，客户端应捕获后对本组重建（1000,2000,…）再重试。
	ErrSortIdOverflow = errors.New("组内排序值已超出上限（65535），请对本组重建后重试")
)
