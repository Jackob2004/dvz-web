package main

import (
	"bytes"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/Jackob2004/dvz-web/internal/database"
	"github.com/Jackob2004/dvz-web/internal/response"
	"github.com/Jackob2004/dvz-web/internal/version"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		app.notFound(w, r)
		return
	}

	data, err := app.constructLeaderboard(0, "level")
	if err != nil {
		app.serverError(w, r, err)
	}

	wrappedData := struct {
		leaderboardData
		Version string
	}{
		data,
		version.Get(),
	}

	app.render(w, r, http.StatusOK, "base", "base", wrappedData)
}

func leaderboardHelper(w http.ResponseWriter, r *http.Request, templateName string, app *application) {
	offset, err := strconv.Atoi(r.PathValue("offset"))
	if err != nil {
		app.badRequest(w, r, errors.New("offset should be an integer"))
		return
	}

	data, err := app.constructLeaderboard(offset, r.PathValue("sort"))
	if err != nil {
		if errors.Is(err, ErrInvalidSortOption) || errors.Is(err, ErrInvalidOffset) {
			app.badRequest(w, r, err)
		} else {
			app.serverError(w, r, err)
		}
		return
	}

	app.render(w, r, http.StatusOK, "home", templateName, data)
}

func (app *application) leaderboardFull(w http.ResponseWriter, r *http.Request) {
	leaderboardHelper(w, r, "leaderboard:full", app)
}

func (app *application) leaderboardComponent(w http.ResponseWriter, r *http.Request) {
	leaderboardHelper(w, r, "leaderboard:update", app)
}

func (app *application) playerHeadSkin(w http.ResponseWriter, r *http.Request) {
	playerUUID := r.PathValue("uuid")

	skin, err := app.db.GetPlayerSkin(playerUUID)
	noRecord := err != nil && errors.Is(err, database.ErrNoRecord)

	if err != nil && !noRecord {
		app.serverError(w, r, err)
		return
	}

	var etag string
	var image []byte
	if noRecord {
		etag = app.skinService.DefaultSkinEtag
		image = app.skinService.DefaultSkin
	} else {
		etag = "\"" + skin.SkinHash + "\""
		image = skin.SkinImage
	}

	w.Header().Set("ETag", etag)
	w.Header().Set("Cache-Control", "public, max-age=28800, must-revalidate")

	if checkETag(r, etag) {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Length", strconv.Itoa(len(image)))
	w.Write(image)
}

func (app *application) playerProfile(w http.ResponseWriter, r *http.Request) {
	playerUUID := r.PathValue("uuid")
	playerProfile, err := app.db.GetWebPlayerProfile(playerUUID)

	if err != nil {
		if errors.Is(err, database.ErrNoRecord) {
			app.notFound(w, r)
		} else {
			app.serverError(w, r, err)
		}
		return
	}

	sort := r.URL.Query().Get("sort")
	nextOffset, err := strconv.Atoi(r.URL.Query().Get("next_offset"))
	prevPageUrl := "/leaderboard/level/0"
	if sort != "" && err == nil {
		prevPageUrl = "/leaderboard/" + sort + "/" + strconv.Itoa(nextOffset-rowsPerPage)
	}

	data := struct {
		database.WebPlayerProfile
		PrevPageUrl string
	}{
		playerProfile,
		prevPageUrl,
	}

	app.render(w, r, http.StatusOK, "player_profile", "profile_page", data)
}

func (app *application) searchPlayer(w http.ResponseWriter, r *http.Request) {
	playerName := r.FormValue("search")
	if playerName == "" {
		app.render(w, r, http.StatusOK, "base", "player-identifier", nil)
		return
	}

	players, err := app.db.GetWebPlayerIdentifiers(playerName)
	if err != nil && !errors.Is(err, database.ErrNoRecord) {
		app.serverError(w, r, err)
		return
	}

	app.render(w, r, http.StatusOK, "base", "player-identifier", players)
}

func (app *application) liveGamePanel(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, http.StatusOK, "livegame", "live-game:page", "none")
}

func (app *application) liveGamePanelUpdate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-store")

	rc := http.NewResponseController(w)
	if err := rc.SetWriteDeadline(time.Time{}); err != nil {
		app.serverError(w, r, err)
		return
	}

	ctx := r.Context()
	channel, subUUID := app.liveGame.subscribe()
	defer app.liveGame.unsubscribe(subUUID)
	buf := new(bytes.Buffer)

	for {
		select {
		case gameInfo := <-channel:
			templateName, err := gameInfo.templateName()
			if err != nil {
				app.logger.Error("liveGamePanelUpdate: error getting template name:", "err", err)
				continue
			}

			err = app.renderSSE(w, buf, gameInfo.unpack(), "livegame", templateName)
			if err != nil {
				if errors.Is(err, response.ErrStreamBroken) {
					app.logger.Debug("liveGamePanelUpdate: error writing to the response: ", "err", err)
					return
				}
				app.logger.Error("liveGamePanelUpdate: error rendering template", "err", err)
			}

		case <-ctx.Done():
			return
		}
	}
}
