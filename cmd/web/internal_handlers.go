package main

import (
	"errors"
	"net/http"

	"github.com/Jackob2004/dvz-web/internal/database"
	"github.com/Jackob2004/dvz-web/internal/request"
	"github.com/Jackob2004/dvz-web/internal/response"
)

func (app *application) ping(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("pong\n"))
}

func (app *application) player(w http.ResponseWriter, r *http.Request) {
	playerUUID := r.PathValue("id")

	playerStats, err := app.db.GetPlayerStatistics(playerUUID)
	if err != nil {
		if errors.Is(err, database.ErrNoRecord) {
			app.notFound(w, r)
		} else {
			app.serverError(w, r, err)
		}
		return
	}

	if err := response.JSON(w, http.StatusOK, playerStats); err != nil {
		app.serverError(w, r, err)
	}
}

func (app *application) playersUpdate(w http.ResponseWriter, r *http.Request) {
	var stats []database.PlayerStatistics
	err := request.DecodeJSON(w, r, &stats)
	if err != nil {
		app.badRequest(w, r, err)
		return
	}

	err = app.db.UpdatePlayersStatistics(stats)

	if err != nil {
		app.serverError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
