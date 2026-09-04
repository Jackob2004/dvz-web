package main

import (
	"net/http"

	"github.com/Jackob2004/dvz-web/internal/response"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		app.notFound(w, r)
		return
	}

	data := app.newTemplateData(r)

	err := response.Page(w, http.StatusOK, data, "pages/home.gohtml")
	if err != nil {
		app.serverError(w, r, err)
	}
}
