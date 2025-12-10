package vo

import "math"

type Pagination struct {
	Total   int64 `json:"total"`
	Page    int   `json:"page"`
	Limit   int   `json:"limit"`
	MaxPage int   `json:"maxPage"`
}

func (p *Pagination) CalcMaxPage() {
	p.MaxPage = int(math.Ceil(float64(p.Total) / float64(p.Limit)))
}
