package view

import (
	"fmt"
	"html/template"
	"path/filepath"

	"github.com/gin-gonic/gin/render"
)

func dict(values ...any) map[string]any {
	result := make(map[string]any, len(values)/2)
	for i := 0; i+1 < len(values); i += 2 {
		key, ok := values[i].(string)
		if ok {
			result[key] = values[i+1]
		}
	}
	return result
}

// Each page has its own template set, so content blocks cannot overwrite other pages.
type Renderer struct{ pages map[string]*template.Template }

func New(root string) (*Renderer, error) {
	pages, err := filepath.Glob(filepath.Join(root, "*", "*.html"))
	if err != nil {
		return nil, err
	}
	r := &Renderer{pages: make(map[string]*template.Template)}
	for _, page := range pages {
		folder := filepath.Base(filepath.Dir(page))
		if folder == "layouts" || folder == "partials" {
			continue
		}
		t, err := template.New("").Funcs(template.FuncMap{"dict": dict, "addOne": func(value int) int { return value + 1 }}).ParseGlob(filepath.Join(root, "layouts", "*.html"))
		if err != nil {
			return nil, fmt.Errorf("parse layouts: %w", err)
		}
		t, err = t.ParseGlob(filepath.Join(root, "partials", "*.html"))
		if err != nil {
			return nil, fmt.Errorf("parse partials: %w", err)
		}
		t, err = t.ParseFiles(page)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", page, err)
		}
		r.pages[folder+"/"+filepath.Base(page)] = t
	}
	if len(r.pages) == 0 {
		return nil, fmt.Errorf("no page templates found in %s", root)
	}
	return r, nil
}

func (r *Renderer) Instance(name string, data any) render.Render {
	t, ok := r.pages[name]
	if !ok {
		panic("unknown page template: " + name)
	}
	return render.HTML{Template: t, Name: "base", Data: data}
}
