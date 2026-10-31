package main

import (
	"errors"

	"github.com/Jackob2004/dvz-web/internal/database"
)

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

const rowsPerPage = 10

var (
	ErrInvalidSortOption = errors.New("invalid sort option")
	ErrInvalidOffset     = errors.New("invalid offset")
)

func (app *application) constructLeaderboard(offset int, sortOption string) (leaderboardData, error) {
	parsedSortOption, err := database.ParseSortOption(sortOption)
	if err != nil {
		return leaderboardData{}, ErrInvalidSortOption
	}

	if offset < 0 || offset%rowsPerPage != 0 {
		return leaderboardData{}, ErrInvalidOffset
	}

	rows, totalRows, err := app.db.WebLeaderboard(rowsPerPage, offset, parsedSortOption)
	if err != nil {
		if errors.Is(err, database.ErrNoRecord) {
			return leaderboardData{
				nil,
				0,
				false,
				false,
				0,
				0,
				0,
				0,
				database.Level,
			}, nil
		}
		return leaderboardData{}, err
	}

	totalPages := totalRows / rowsPerPage
	if totalRows%rowsPerPage != 0 {
		totalPages++
	}
	currPage := offset/rowsPerPage + 1

	return leaderboardData{
		rows,
		rowsPerPage - len(rows),
		currPage < totalPages,
		currPage != 1,
		currPage,
		totalPages,
		offset + rowsPerPage,
		offset - rowsPerPage,
		parsedSortOption,
	}, nil
}
