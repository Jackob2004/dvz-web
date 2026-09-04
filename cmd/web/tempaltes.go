package main

import (
	"net/http"

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
