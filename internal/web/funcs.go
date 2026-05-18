package web

import (
	"fmt"
	"html/template"
	"time"
)

func funcMap() template.FuncMap {
	return template.FuncMap{
		"formatTime": func(t time.Time) string {
			if t.IsZero() {
				return "—"
			}
			return t.Local().Format("2006-01-02 15:04")
		},
		"formatTimePtr": func(t *time.Time) string {
			if t == nil || t.IsZero() {
				return "—"
			}
			return t.Local().Format("2006-01-02 15:04")
		},
		"formatDate": func(t time.Time) string {
			return t.Local().Format("2006-01-02")
		},
		"hours": func(h float64) string {
			return fmt.Sprintf("%.1f", h)
		},
		"statusBadge": func(status string) template.HTML {
			cls := "bg-slate-100 text-slate-600 border border-slate-200"
			label := status
			switch status {
			case "present":
				cls = "bg-emerald-50 text-emerald-700 border border-emerald-200"
				label = "присутствует"
			case "late":
				cls = "bg-amber-50 text-amber-700 border border-amber-200"
				label = "опоздание"
			case "absent":
				cls = "bg-rose-50 text-rose-700 border border-rose-200"
				label = "отсутствует"
			case "pending":
				cls = "bg-blue-50 text-blue-700 border border-blue-200"
				label = "ожидает"
			case "approved":
				cls = "bg-emerald-50 text-emerald-700 border border-emerald-200"
				label = "одобрено"
			case "declined":
				cls = "bg-rose-50 text-rose-700 border border-rose-200"
				label = "отклонено"
			}
			return template.HTML(fmt.Sprintf(
				`<span class="inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium %s">%s</span>`,
				cls, template.HTMLEscapeString(label),
			))
		},
		"roleLabel": func(role string) string {
			switch role {
			case "admin":
				return "Администратор"
			case "manager":
				return "Руководитель"
			case "employee":
				return "Сотрудник"
			}
			return role
		},
		"statusLabel": func(status string) string {
			switch status {
			case "pending":
				return "ожидающие"
			case "approved":
				return "одобренные"
			case "declined":
				return "отклонённые"
			case "":
				return "все"
			}
			return status
		},
		"derefInt": func(p *int) int {
			if p == nil {
				return 0
			}
			return *p
		},
		"derefStr": func(p *string) string {
			if p == nil {
				return ""
			}
			return *p
		},
		"today": func() string {
			return time.Now().Local().Format("2006-01-02")
		},
		"strs": func(values ...string) []string { return values },
		"firstChar": func(s string) string {
			for _, r := range s {
				return string(r)
			}
			return ""
		},
		"dict": func(values ...any) (map[string]any, error) {
			if len(values)%2 != 0 {
				return nil, fmt.Errorf("dict requires even arg count")
			}
			m := make(map[string]any, len(values)/2)
			for i := 0; i < len(values); i += 2 {
				key, ok := values[i].(string)
				if !ok {
					return nil, fmt.Errorf("dict keys must be strings")
				}
				m[key] = values[i+1]
			}
			return m, nil
		},
	}
}
