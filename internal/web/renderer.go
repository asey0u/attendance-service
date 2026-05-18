package web

import (
	"bytes"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"path"
	"strings"
)

type Renderer struct {
	pages    map[string]*template.Template
	partials *template.Template
}

func NewRenderer() (*Renderer, error) {
	r := &Renderer{pages: map[string]*template.Template{}}

	partials, err := template.New("partials").
		Funcs(funcMap()).
		ParseFS(templatesFS, "templates/partials/*.html")
	if err != nil {
		return nil, fmt.Errorf("parse partials: %w", err)
	}
	r.partials = partials

	pageEntries, err := fs.ReadDir(templatesFS, "templates/pages")
	if err != nil {
		return nil, fmt.Errorf("read pages dir: %w", err)
	}

	for _, e := range pageEntries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".html") {
			continue
		}
		name := strings.TrimSuffix(e.Name(), ".html")

		t, err := template.New(name).Funcs(funcMap()).ParseFS(
			templatesFS,
			"templates/partials/*.html",
			"templates/layout/base.html",
			path.Join("templates/pages", e.Name()),
		)
		if err != nil {
			return nil, fmt.Errorf("parse page %s: %w", name, err)
		}
		r.pages[name] = t
	}

	return r, nil
}

func (r *Renderer) Page(w http.ResponseWriter, status int, name string, data any) {
	t, ok := r.pages[name]
	if !ok {
		http.Error(w, "template not found: "+name, http.StatusInternalServerError)
		return
	}
	var buf bytes.Buffer
	if err := t.ExecuteTemplate(&buf, "base", data); err != nil {
		http.Error(w, "render: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, _ = buf.WriteTo(w)
}

func (r *Renderer) Partial(w http.ResponseWriter, name string, data any) {
	var buf bytes.Buffer
	if err := r.partials.ExecuteTemplate(&buf, name, data); err != nil {
		http.Error(w, "render: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = buf.WriteTo(w)
}
