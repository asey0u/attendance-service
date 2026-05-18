package httpx

import (
	"net/http"
	"strconv"
)

const PageSize = 20

type Pager struct {
	Page       int
	PageSize   int
	Total      int
	TotalPages int
	BaseURL    string // path + all params except page=, ending with "?" or "&"
}

func NewPager(page, total int, baseURL string) Pager {
	totalPages := (total + PageSize - 1) / PageSize
	if totalPages < 1 {
		totalPages = 1
	}
	if page < 1 {
		page = 1
	}
	if page > totalPages {
		page = totalPages
	}
	return Pager{
		Page:       page,
		PageSize:   PageSize,
		Total:      total,
		TotalPages: totalPages,
		BaseURL:    baseURL,
	}
}

func ParsePage(r *http.Request) int {
	p, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if p < 1 {
		return 1
	}
	return p
}

func BaseURLFrom(r *http.Request) string {
	q := r.URL.Query()
	q.Del("page")
	if len(q) == 0 {
		return r.URL.Path + "?"
	}
	return r.URL.Path + "?" + q.Encode() + "&"
}

func (p Pager) Offset() int     { return (p.Page - 1) * p.PageSize }
func (p Pager) HasPrev() bool   { return p.Page > 1 }
func (p Pager) HasNext() bool   { return p.Page < p.TotalPages }
func (p Pager) PrevPage() int   { return p.Page - 1 }
func (p Pager) NextPage() int   { return p.Page + 1 }

func (p Pager) From() int {
	if p.Total == 0 {
		return 0
	}
	return p.Offset() + 1
}

func (p Pager) To() int {
	t := p.Page * p.PageSize
	if t > p.Total {
		return p.Total
	}
	return t
}
