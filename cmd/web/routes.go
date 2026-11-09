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
	mux.HandleFunc("GET /leaderboard/{sort}/{offset}", app.leaderboardFull)
	mux.HandleFunc("GET /leaderboard-component/{sort}/{offset}", app.leaderboardComponent)
	mux.HandleFunc("GET /player/skin/head/{uuid}", app.playerHeadSkin)
	mux.HandleFunc("GET /player/profile/{uuid}", app.playerProfile)
	mux.HandleFunc("GET /livegame", app.liveGamePanel)
	mux.HandleFunc("GET /livegame/events", app.liveGamePanelUpdate)

	return app.logRequest(app.recoverPanic(app.securityHeaders(mux)))
}

func (app *application) internalRoutes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/ping", app.ping)
	mux.HandleFunc("GET /api/player/{uuid}", app.player)
	mux.HandleFunc("GET /api/players/leaderboard", app.playersLeaderboard)
	mux.HandleFunc("PUT /api/players/update", app.playersUpdate)
	mux.HandleFunc("PUT /api/livegame/update", app.liveGameUpdate)

	return app.logRequest(app.recoverPanic(mux))
}
