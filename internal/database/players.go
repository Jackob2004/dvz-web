package database

import (
	"database/sql"
	"errors"
)

// PlayerStatistics Represents data that is collected while DvZ plugin game match is on
type PlayerStatistics struct {
	UUID                string `db:"uuid" json:"uuid"`
	Name                string `db:"username" json:"username"`
	ExperiencePoints    int    `db:"experience_points" json:"experience_points"`
	DwarfKills          int    `db:"dwarf_kills" json:"dwarf_kills"`
	ZombieKills         int    `db:"zombie_kills" json:"zombie_kills"`
	AiZombieKills       int    `db:"ai_zombie_kills" json:"ai_zombie_kills"`
	ShrineKills         int    `db:"shrine_kills" json:"shrine_kills"`
	Deaths              int    `db:"deaths" json:"deaths"`
	LongestTimeSurvived int64  `db:"longest_time_survived" json:"longest_time_survived_seconds"`
	TotalPlayTime       int64  `db:"total_play_time" json:"total_play_time_seconds"`
	GamesPlayed         int    `db:"games_played" json:"games_played"`
	RampageKillStreak   int    `db:"rampage_kill_streak" json:"rampage_kill_streak"`
	BowShoots           int    `db:"bow_shoots" json:"bow_shoots"`
	BowHits             int    `db:"bow_hits" json:"bow_hits"`
}

type LeaderboardRow struct {
	Name  string `db:"username" json:"username"`
	Level int    `db:"level" json:"level"`
}

func (db *DB) GetPlayerStatistics(playerUUID string) (PlayerStatistics, error) {
	var stats PlayerStatistics

	stmt := `SELECT uuid, username, experience_points, dwarf_kills, zombie_kills, ai_zombie_kills, shrine_kills, deaths,
       longest_time_survived, total_play_time, games_played, rampage_kill_streak, bow_shoots, bow_hits
	FROM players WHERE uuid = ?`

	err := db.Get(&stats, stmt, playerUUID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return PlayerStatistics{}, ErrNoRecord
		}
		return PlayerStatistics{}, err
	}

	return stats, nil
}

func (db *DB) UpdatePlayersStatistics(stats []PlayerStatistics) error {
	tx, err := db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, stat := range stats {
		stmt := `INSERT INTO players 
		(uuid, username, experience_points, dwarf_kills, zombie_kills, ai_zombie_kills, shrine_kills, deaths, longest_time_survived, total_play_time, games_played, rampage_kill_streak, bow_shoots, bow_hits)
		VALUES (:uuid, :username, :experience_points, :dwarf_kills, :zombie_kills, :ai_zombie_kills, :shrine_kills, :deaths, :longest_time_survived, :total_play_time, :games_played, :rampage_kill_streak, :bow_shoots, :bow_hits)
		ON CONFLICT (uuid) DO UPDATE SET
			username = excluded.username,
			experience_points = excluded.experience_points,
			dwarf_kills = excluded.dwarf_kills,
			zombie_kills = excluded.zombie_kills,
			ai_zombie_kills = excluded.ai_zombie_kills,
			shrine_kills = excluded.shrine_kills,
			deaths = excluded.deaths,
			longest_time_survived = excluded.longest_time_survived,
			total_play_time = excluded.total_play_time,
			games_played = excluded.games_played,
			rampage_kill_streak = excluded.rampage_kill_streak,
			bow_shoots = excluded.bow_shoots,
			bow_hits = excluded.bow_hits
		`
		_, err = tx.NamedExec(stmt, stat)
		if err != nil {
			return err
		}
	}

	err = tx.Commit()
	return err
}

func (db *DB) Leaderboard() ([]LeaderboardRow, error) {
	var leaderboardRows []LeaderboardRow
	stmt := "SELECT username, level FROM players ORDER BY level DESC LIMIT 10"

	err := db.Select(&leaderboardRows, stmt)
	if err != nil {
		return nil, err
	}
	
	if len(leaderboardRows) == 0 {
		return nil, ErrNoRecord
	}

	return leaderboardRows, nil
}
