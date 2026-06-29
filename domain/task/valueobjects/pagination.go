package valueobjects

import "math"

// Pagination 分页值对象
type Pagination struct {
	Total   int64
	Page    int
	Limit   int
	MaxPage int
}

// CalcMaxPage 计算最大页数
func (p *Pagination) CalcMaxPage() {
	p.MaxPage = int(math.Ceil(float64(p.Total) / float64(p.Limit)))
}

// NewPagination 创建分页值对象
func NewPagination(total int64, page int, limit int) *Pagination {
	return &Pagination{
		Total:   total,
		Page:    page,
		Limit:   limit,
		MaxPage: 0,
	}
}
