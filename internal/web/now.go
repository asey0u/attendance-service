package web

import (
	"net/http"
	"time"
)

var nowFunc = time.Now

func parseDateParam(r *http.Request, key string, addDay bool) *time.Time {
	s := r.URL.Query().Get(key)
	if s == "" {
		return nil
	}
	t, err := time.ParseInLocation("2006-01-02", s, time.Local)
	if err != nil {
		return nil
	}
	if addDay {
		t = t.Add(24*time.Hour - time.Second)
	}
	return &t
}
