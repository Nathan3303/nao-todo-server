package controllers

import (
	"fmt"
	"net/http"

	"naotodoserver/infrastructure/utils/query"
	"naotodoserver/interfaces/types"

	"github.com/gin-gonic/gin"
)

// normalizePagination 归一化列表请求的 page/limit（SEC-01 / T561）。
//   - page > query.MaxPage ⇒ HTTP 400（深分页会长时间占用连接池）
//   - limit > query.MaxLimit ⇒ 夹取到上限（向后兼容：不拒绝合法请求）
//
// 成功返回 true 并就地改写入参；失败已写 400 响应，调用方直接 return 即可。
// @param ctx gin 上下文
// @param page 页码指针（就地改写）
// @param limit 每页条数指针（就地改写）
// @param code 业务错误码（沿用各控制器既有的「参数错误」码）
func normalizePagination(ctx *gin.Context, page, limit *int, code int) bool {
	normalizedPage, err := query.NormalizePage(*page)
	if err != nil {
		FailureByHttpStatus(ctx, http.StatusBadRequest, types.ResponseData{
			Code:    code,
			Message: "参数错误",
			Error:   fmt.Sprintf("page 超出上限（最大 %d）", query.MaxPage),
		})
		return false
	}
	*page = normalizedPage
	*limit = query.ClampLimit(*limit)
	return true
}
