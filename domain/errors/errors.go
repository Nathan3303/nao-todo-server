package errors

import (
	"errors"
	"fmt"
)

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
	// ErrLoginThrottled 账号因连续登录失败被临时暂停（SEC-04）；
	// 对外响应必须与 ErrPasswordMismatch 完全一致，避免账号枚举。
	ErrLoginThrottled = errors.New("登录尝试过于频繁")
	// ErrPasswordTooWeak 口令不满足密码策略（SEC-05）；
	// 对外只回统一可读文案，不回显具体命中了哪条规则。
	ErrPasswordTooWeak = errors.New("密码不符合安全策略")
)

// PublicError 可安全回显给客户端的业务错误（SEC-12）。
// 文案必须由服务端显式声明：不得包含 SQL、表名/列名、主键值、文件路径。
// 接口层据 IsPublicError 判定"可回显"，其余一律用通用文案替代并落日志。
type PublicError struct {
	Message string
}

// Error 实现 error 接口
func (e *PublicError) Error() string {
	return e.Message
}

// Publicf 构造可回显的业务错误（支持格式化参数，便于保留"最多拥有 %d 个附件"等原文案）
// @param format 文案模板
// @param args 格式化参数
// @return 可回显的业务错误
func Publicf(format string, args ...any) error {
	return &PublicError{Message: fmt.Sprintf(format, args...)}
}

// IsPublicError 判断错误是否可安全回显（含被 %w 包装的情况）
// @param err 任意错误
// @return 可回显返回 true
func IsPublicError(err error) bool {
	var publicErr *PublicError
	return errors.As(err, &publicErr)
}
