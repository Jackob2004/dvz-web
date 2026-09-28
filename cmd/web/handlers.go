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
	}

	err := response.Page(w, http.StatusOK, data, "pages/home.gohtml")
	if err != nil {
		app.serverError(w, r, err)
	}
}

const rowsPerPage = 10

func (app *application) leaderboardComponent(w http.ResponseWriter, r *http.Request) {
	offset, err := strconv.Atoi(r.PathValue("offset"))
	if err != nil || offset < 0 || offset%rowsPerPage != 0 {
		app.badRequest(w, r, errors.New("invalid offset"))
		return
	}

	rows, totalRows, err := app.db.WebLeaderboard(rowsPerPage, offset)
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
	}

	err = response.NamedTemplate(w, http.StatusOK, data, "leaderboard", "pages/home.gohtml")
	if err != nil {
		app.serverError(w, r, err)
	}
}
