package controllers

import (
	"errors"

	domerr "naotodoserver/domain/errors"
	iCtx "naotodoserver/infrastructure/context"
	"naotodoserver/infrastructure/logging"
	"naotodoserver/interfaces/types"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// 对外白名单文案（SEC-12）：(b) 存在性 / (c) 基础设施
const (
	// publicMsgNotFound 资源不存在/不属于本租户/已删除三态统一文案（消除存在性 oracle）
	publicMsgNotFound = "记录不存在"
	// publicMsgInternal 基础设施类失败统一文案（原始错误只进日志）
	publicMsgInternal = "服务暂时不可用，请稍后重试"
)

// publicErrSentinels 可安全回显的领域哨兵（我们自己的中文文案，不含内部细节）
var publicErrSentinels = []error{
	domerr.ErrInvalidUserID,
	domerr.ErrInvalidID,
	domerr.ErrInvalidProjectID,
	domerr.ErrInvalidTaskID,
	domerr.ErrInvalidTagID,
	domerr.ErrInvalidCommentID,
	domerr.ErrInvalidItemID,
	domerr.ErrInvalidPomodoroID,
	domerr.ErrPasswordMismatch,
	domerr.ErrPasswordTooWeak,
	domerr.ErrLoginThrottled,
	domerr.ErrEmailExists,
	domerr.ErrUserDeactivated,
	domerr.ErrUserInCooldown,
	domerr.ErrCreateFailed,
	domerr.ErrDeleteFailed,
	domerr.ErrIDConflict,
	domerr.ErrSortIdOverflow,
}

// notFoundSentinels (b) 类资源存在性哨兵
var notFoundSentinels = []error{
	domerr.ErrProjectNotFound,
	domerr.ErrUserNotFound,
	domerr.ErrSessionNotFound,
}

// failWithErr 统一的失败出口（保留端点业务文案，Error 字段只放白名单文案）
// @param ctx 请求上下文
// @param code 既有业务码（保持不变）
// @param userMsg 端点业务文案（如「获取清单失败」）
// @param err 原始错误（只用于分类；基础设施错误只进日志）
func failWithErr(ctx *gin.Context, code int, userMsg string, err error) {
	var text string
	switch {
	case err == nil:
		text = ""
	case isNotFoundErr(err):
		logServerErr(ctx, code, err)
		text = publicMsgNotFound
	default:
		publicText, ok := publicTextOf(err)
		if ok {
			// 只用规范文案：err.Error() 可能带 "包.函数: " 的 wrap 前缀（F1）
			text = publicText
			break
		}
		logServerErr(ctx, code, err)
		text = publicMsgInternal
	}
	Failure(ctx, types.ResponseData{
		Code:    code,
		Message: userMsg,
		Error:   text,
	})
}

// failWithErrMessage 用于原本把 err.Error() 直接当 Message 回显的端点：
// 可读业务错误保持原文案（向后兼容），其余改为端点固定文案 + 白名单 Error。
// @param ctx 请求上下文
// @param code 既有业务码（保持不变）
// @param fallbackMsg 基础设施/未知失败时的固定文案
// @param err 原始错误
func failWithErrMessage(ctx *gin.Context, code int, fallbackMsg string, err error) {
	if publicText, ok := publicTextOf(err); ok {
		Failure(ctx, types.ResponseData{
			Code:    code,
			Message: publicText,
		})
		return
	}
	failWithErr(ctx, code, fallbackMsg, err)
}

// failBadRequestBind 参数绑定/读取失败：保留既有「参数错误」类文案，
// 原始 binder 细节只进日志（不再回显字段名/解析位置）。
// @param ctx 请求上下文
// @param code 既有业务码（保持不变）
// @param userMsg 既有参数错误文案
// @param err 绑定错误
func failBadRequestBind(ctx *gin.Context, code int, userMsg string, err error) {
	logServerErr(ctx, code, err)
	Failure(ctx, types.ResponseData{
		Code:    code,
		Message: userMsg,
	})
}

// publicErrText 计算可回显文案（不写日志，供内部组合使用）
// @param err 原始错误
// @return 白名单文案
func publicErrText(err error) string {
	switch {
	case err == nil:
		return ""
	case isNotFoundErr(err):
		return publicMsgNotFound
	default:
		if publicText, ok := publicTextOf(err); ok {
			return publicText
		}
		return publicMsgInternal
	}
}

// publicTextOf 返回可显式回显的规范文案（不含 wrap 前缀），
// 用于避免把 fmt.Errorf("pkg.Fn: %w", ...) 的内部前缀回给客户端（F1）。
// @param err 原始错误（可被多层 %w 包装）
// @return 规范文案 / 是否可回显
func publicTextOf(err error) (string, bool) {
	if err == nil {
		return "", false
	}
	var publicErr *domerr.PublicError
	if errors.As(err, &publicErr) {
		return publicErr.Message, true
	}
	for _, sentinel := range publicErrSentinels {
		if errors.Is(err, sentinel) {
			return sentinel.Error(), true
		}
	}
	return "", false
}

// isPublicErr 判断错误文案是否可安全回显：
// 显式构造的业务错误（PublicError）或白名单领域哨兵
// @param err 原始错误
// @return 可回显返回 true
func isPublicErr(err error) bool {
	_, ok := publicTextOf(err)
	return ok
}

// isNotFoundErr 判断是否为资源存在性错误（含 ORM 的 record not found）
// @param err 原始错误
// @return 命中返回 true
func isNotFoundErr(err error) bool {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return true
	}
	for _, sentinel := range notFoundSentinels {
		if errors.Is(err, sentinel) {
			return true
		}
	}
	return false
}

// logServerErr 记录不可回显的内部错误（含请求关联标识，便于按 request_id 追溯）
// @param ctx 请求上下文
// @param code 业务码（用于定位端点分支）
// @param err 原始错误
func logServerErr(ctx *gin.Context, code int, err error) {
	if err == nil {
		return
	}
	requestCtx := ctx.Request.Context()
	fields := map[string]any{
		"event":      "api.error",
		"request_id": iCtx.GetRequestId(requestCtx),
		"code":       code,
		"method":     ctx.Request.Method,
		"path":       ctx.Request.URL.Path,
		"error":      err.Error(),
	}
	if userId := iCtx.GetUserId(requestCtx); userId > 0 {
		fields["user_id"] = userId
	}
	logging.ErrorWithFields(fields, "接口处理失败（原始错误仅记录在服务端）")
}
