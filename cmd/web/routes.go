package main

import (
	"net/http"

	"github.com/Jackob2004/dvz-web/assets"
)

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	fileServer := http.FileServer(http.FS(assets.EmbeddedFiles))
	mux.Handle("GET /static/", fileServer)

	mux.HandleFunc("GET /", app.home)

	return app.logRequest(app.recoverPanic(app.securityHeaders(mux)))
}

func (app *application) internalRoutes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /ping", app.ping)

	return app.logRequest(app.recoverPanic(mux))
}
