package query

import (
	"errors"
	"testing"
)

// TestClampLimit limit 上限夹取（SEC-01）：超限夹到 500，<=0 原样返回
func TestClampLimit(t *testing.T) {
	cases := []struct {
		in   int
		want int
	}{
		{100000, MaxLimit},
		{501, MaxLimit},
		{500, 500},
		{10, 10},
		{0, 0}, // 默认值由调用方补齐，保持既有语义
		{-1, -1},
	}
	for _, c := range cases {
		if got := ClampLimit(c.in); got != c.want {
			t.Fatalf("ClampLimit(%d) = %d, want %d", c.in, got, c.want)
		}
	}
}

// TestNormalizePage 深分页拒绝（SEC-01）：<=0 ⇒ 1，> MaxPage ⇒ ErrPageTooDeep
func TestNormalizePage(t *testing.T) {
	if got, err := NormalizePage(0); err != nil || got != 1 {
		t.Fatalf("NormalizePage(0) = %d/%v, want 1/nil", got, err)
	}
	if got, err := NormalizePage(MaxPage); err != nil || got != MaxPage {
		t.Fatalf("NormalizePage(MaxPage) = %d/%v, want %d/nil", got, err, MaxPage)
	}
	if _, err := NormalizePage(MaxPage + 1); !errors.Is(err, ErrPageTooDeep) {
		t.Fatalf("NormalizePage(MaxPage+1) err = %v, want ErrPageTooDeep", err)
	}
}

// TestPaginationWindow Paginate 兜底窗口：超大 limit ⇒ 500；超深 page ⇒ MaxPage
func TestPaginationWindow(t *testing.T) {
	size, offset := paginationWindow(1, 100000)
	if size != MaxLimit || offset != 0 {
		t.Fatalf("window(1,100000) = %d/%d, want %d/0", size, offset, MaxLimit)
	}

	size, offset = paginationWindow(MaxPage+1, 10)
	if size != 10 || offset != (MaxPage-1)*10 {
		t.Fatalf("window(MaxPage+1,10) = %d/%d, want 10/%d", size, offset, (MaxPage-1)*10)
	}

	size, offset = paginationWindow(0, 0)
	if size != DefaultLimit || offset != 0 {
		t.Fatalf("window(0,0) = %d/%d, want %d/0", size, offset, DefaultLimit)
	}
}
