package query

import (
	"errors"

	"gorm.io/gorm"
)

// 分页上限（SEC-01 / T561）：limit 无上限与深分页会长时间占用连接池（SetMaxOpenConns(30)），
// 表现为全站延迟/超时。上限在接口层拒绝或夹取，这里做最后一道兜底。
const (
	// DefaultLimit limit 缺省值（沿用既有语义）
	DefaultLimit = 10
	// MaxLimit 单页/单次同步最大条数
	MaxLimit = 500
	// MaxPage 最大页码；page > MaxPage ⇒ 深分页拒绝（接口层返回 400）
	MaxPage = 10000
)

// ErrPageTooDeep 深分页拒绝；接口层映射为 400（可读提示，不透传 DB 错误）
var ErrPageTooDeep = errors.New("page 超出上限")

// ClampLimit 夹取 limit 上限（≤MaxLimit）。
// <=0 原样返回：默认值由调用方补齐，保持既有默认值语义（列表 10 / 同步 100）。
// @param limit 请求传入的每页条数
// @return 夹取后的每页条数
func ClampLimit(limit int) int {
	if limit > MaxLimit {
		return MaxLimit
	}
	return limit
}

// NormalizePage 归一化页码：<=0 ⇒ 1；> MaxPage ⇒ ErrPageTooDeep。
// @param page 请求传入的页码
// @return 归一化后的页码
// @return error 深分页时为 ErrPageTooDeep
func NormalizePage(page int) (int, error) {
	if page <= 0 {
		return 1, nil
	}
	if page > MaxPage {
		return 0, ErrPageTooDeep
	}
	return page, nil
}

// Paginate 分页 Scope
// 通用分页逻辑，从 PaginationVO2Scopes 迁移并通用化；
// SEC-01 起在此兜底 limit 上限与页码上限（接口层已先行拒绝深分页）。
func Paginate(page, limit int) func(db *gorm.DB) *gorm.DB {
	size, offset := paginationWindow(page, limit)
	return func(db *gorm.DB) *gorm.DB {
		return db.Offset(offset).Limit(size)
	}
}

// paginationWindow 归一化分页窗口（纯逻辑，便于单测）：
// limit <=0 ⇒ DefaultLimit，limit > MaxLimit ⇒ MaxLimit；
// page 复用 NormalizePage，深分页（接口层已 400）在此兜底到 MaxPage。
// @return 每页条数、偏移量
func paginationWindow(page, limit int) (int, int) {
	normalizedPage, err := NormalizePage(page)
	if err != nil {
		normalizedPage = MaxPage
	}
	if limit <= 0 {
		limit = DefaultLimit
	}
	limit = ClampLimit(limit)
	return limit, (normalizedPage - 1) * limit
}
