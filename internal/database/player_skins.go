package database

import (
	"database/sql"
	"errors"
)

type PlayerSkinData struct {
	PlayerID  int64   `db:"id"`
	SkinHash  *string `db:"skin_hash"`
	FetchedAt *int64  `db:"fetched_at"`
}

type PlayerSkin struct {
	SkinImage []byte `db:"head_skin"`
	SkinHash  string `db:"skin_hash"`
}

func (db *DB) GetPlayerSkin(playerUUID string) (PlayerSkin, error) {
	var skinImage PlayerSkin

	stmt := `SELECT head_skin, skin_hash FROM player_skins
    WHERE player_id = (SELECT id FROM players WHERE uuid = ?)`

	err := db.Get(&skinImage, stmt, playerUUID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return PlayerSkin{}, ErrNoRecord
		}
		return PlayerSkin{}, err
	}

	return skinImage, err
}

func (db *DB) GetPlayerSkinData(playerUUID string) (PlayerSkinData, error) {
	var skinData PlayerSkinData

	stmt := `SELECT players.id, player_skins.skin_hash, player_skins.fetched_at FROM players
    LEFT JOIN player_skins ON player_skins.player_id = players.id
    WHERE players.uuid = ?`

	err := db.Get(&skinData, stmt, playerUUID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return PlayerSkinData{}, ErrNoRecord
		}
		return PlayerSkinData{}, err
	}

	return skinData, err
}

func (db *DB) UpdatePlayerSkin(playerID int64, skinHash string, skin []byte, fetchedAt int64) error {
	stmt := `INSERT INTO player_skins (player_id, head_skin, skin_hash, fetched_at)
	VALUES (?, ?, ?, ?)
	ON CONFLICT (player_id) DO UPDATE SET
	head_skin = EXCLUDED.head_skin,
	skin_hash = EXCLUDED.skin_hash,
	fetched_at = EXCLUDED.fetched_at
	`

	_, err := db.Exec(stmt, playerID, skin, skinHash, fetchedAt)
	if err != nil {
		return err
	}

	return nil
}
