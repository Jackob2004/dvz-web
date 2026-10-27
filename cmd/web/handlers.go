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

	data := leaderboardData{
		nil,
		10,
		false,
		false,
		0,
		0,
		0,
		0,
		database.Level,
	}

	err := response.Page(w, http.StatusOK, data, "pages/home.gohtml")
	if err != nil {
		app.serverError(w, r, err)
	}
}

const rowsPerPage = 10

func (app *application) leaderboardComponent(w http.ResponseWriter, r *http.Request) {
	sortOption, err := database.ParseSortOption(r.PathValue("sort"))
	if err != nil {
		app.badRequest(w, r, err)
		return
	}

	offset, err := strconv.Atoi(r.PathValue("offset"))
	if err != nil || offset < 0 || offset%rowsPerPage != 0 {
		app.badRequest(w, r, errors.New("invalid offset"))
		return
	}

	rows, totalRows, err := app.db.WebLeaderboard(rowsPerPage, offset, sortOption)
	if err != nil {
		if errors.Is(err, database.ErrNoRecord) {
			data := leaderboardData{
				nil,
				0,
				false,
				false,
				0,
				0,
				0,
				0,
				database.Level,
			}
			err := response.NamedTemplate(w, http.StatusOK, data, "leaderboard", "pages/home.gohtml")
			if err != nil {
				app.serverError(w, r, err)
			}
		} else {
			app.serverError(w, r, err)
		}
		return
	}

	totalPages := totalRows / rowsPerPage
	if totalRows%rowsPerPage != 0 {
		totalPages++
	}
	currPage := offset/rowsPerPage + 1

	data := leaderboardData{
		rows,
		rowsPerPage - len(rows),
		currPage < totalPages,
		currPage != 1,
		currPage,
		totalPages,
		offset + rowsPerPage,
		offset - rowsPerPage,
		sortOption,
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	err = response.NamedTemplate(w, http.StatusOK, data, "leaderboard", "pages/home.gohtml")
	if err != nil {
		app.serverError(w, r, err)
	}
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

	err = response.NamedTemplate(w, http.StatusOK, playerProfile, "profile_page", "pages/player_profile.gohtml")
	if err != nil {
		app.serverError(w, r, err)
	}
}
