package main

import (
	"net/http"

	"github.com/Jackob2004/dvz-web/internal/database"
	"github.com/Jackob2004/dvz-web/internal/version"
)

type templateData struct {
	Version string
}

type leaderboardData struct {
	PlayerRows      []database.WebLeaderboardRow
	PlaceholderRows int
	HasNext         bool
	HasPrev         bool
	CurrPage        int
	TotalPages      int
	NextOffset      int
	PrevOffset      int
	SortingOption   database.SortOption
}

func (app *application) newTemplateData(r *http.Request) templateData {
	return templateData{
		Version: version.Get(),
	}
}
