package main

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/Jackob2004/dvz-web/internal/database"
	"github.com/Jackob2004/dvz-web/internal/response"
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

	err = response.Page(w, http.StatusOK, data, "pages/home.gohtml")
	if err != nil {
		app.serverError(w, r, err)
	}
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

	err = response.NamedTemplate(w, http.StatusOK, data, templateName, "pages/home.gohtml")
	if err != nil {
		app.serverError(w, r, err)
	}
}

func (app *application) leaderboardFull(w http.ResponseWriter, r *http.Request) {
	leaderboardHelper(w, r, "leaderboard:full", app)
}

func (app *application) leaderboardComponent(w http.ResponseWriter, r *http.Request) {
	leaderboardHelper(w, r, "leaderboard:update", app)
}

func (app *application) playerHeadSkin(w http.ResponseWriter, r *http.Request) {
	playerUUID := r.PathValue("id")

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
	playerUUID := r.PathValue("id")
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

	err = response.NamedTemplate(w, http.StatusOK, data, "profile_page", "pages/player_profile.gohtml")
	if err != nil {
		app.serverError(w, r, err)
	}
}
