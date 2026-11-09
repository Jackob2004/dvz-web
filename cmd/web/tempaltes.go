package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/Jackob2004/dvz-web/assets"
	"github.com/Jackob2004/dvz-web/internal/funcs"
	"github.com/Jackob2004/dvz-web/internal/response"
	"github.com/Jackob2004/dvz-web/internal/version"
)

type templateData struct {
	Version string
}

func (app *application) newTemplateData(r *http.Request) templateData {
	return templateData{
		Version: version.Get(),
	}
}

func (app *application) render(w http.ResponseWriter, r *http.Request, status int, page, templateName string, data any) {
	t, ok := app.templateCache[page]
	if !ok {
		app.serverError(w, r, fmt.Errorf("page %s not found in template cache", page))
		return
	}

	buf := new(bytes.Buffer)
	err := t.ExecuteTemplate(buf, templateName, data)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	w.WriteHeader(status)
	buf.WriteTo(w)
}

func (app *application) renderSSE(w http.ResponseWriter, buffer *bytes.Buffer, data any, page, templateName string) error {
	t, ok := app.templateCache[page]
	if !ok {
		return fmt.Errorf("page %s not found in template cache", page)
	}

	return response.NamedSSETemplate(w, buffer, data, t, templateName)
}

func newTemplateCache() (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

	basePatterns := []string{"templates/base.gohtml", "templates/partials/*.gohtml", "templates/pages/home.gohtml"}

	ts, err := template.New("").Funcs(funcs.TemplateFuncs).ParseFS(assets.EmbeddedFiles, basePatterns...)
	if err != nil {
		return nil, err
	}

	cache["base"] = ts

	pages, err := fs.Glob(assets.EmbeddedFiles, "templates/pages/*.gohtml")
	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		name, _ := strings.CutSuffix(filepath.Base(page), ".gohtml")

		ts, err = template.New("").Funcs(funcs.TemplateFuncs).ParseFS(assets.EmbeddedFiles, page)
		if err != nil {
			return nil, err
		}

		cache[name] = ts
	}

	return cache, nil
}
