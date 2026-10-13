package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"io/fs"
	"net/http"
	"time"

	"github.com/Jackob2004/dvz-web/assets"
	"github.com/Jackob2004/dvz-web/internal/database"
)

var defaultSkin []byte
var defaultSkinETag string

var customClient = &http.Client{Timeout: 5 * time.Second}

func hashImage(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

func (app *application) initDefaultSkin() {
	data, err := fs.ReadFile(assets.EmbeddedFiles, "static/img/skin_head_default.png")
	if err != nil {
		app.logger.Error("error reading embedded images: %v", err)
	}

	defaultSkin = data
	defaultSkinETag = "\"" + hashImage(data) + "\""
}

func fetchPlayerHeadSkin(playerName string) ([]byte, error) {
	resp, err := customClient.Get("https://api.mcheads.org/head/" + playerName + "/64")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("request failed response code:" + resp.Status)
	}

	imageBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return imageBytes, nil
}

func saveNewPlayerSkin(db *database.DB, skinData database.PlayerSkinData, playerName string) error {
	headSkin, err := fetchPlayerHeadSkin(playerName)
	if err != nil {
		return err
	}

	return db.UpdatePlayerSkin(skinData.PlayerID, hashImage(headSkin), headSkin, time.Now().Unix())
}

func updateExistingPlayerSkin(db *database.DB, skinData database.PlayerSkinData, playerName string) error {
	oneWeek := time.Hour * 167
	if time.Since(time.Unix(*skinData.FetchedAt, 0)) < oneWeek {
		return nil
	}

	headSkin, err := fetchPlayerHeadSkin(playerName)
	if err != nil {
		return err
	}

	newHash := hashImage(headSkin)
	if newHash == *skinData.SkinHash {
		return nil
	}

	return db.UpdatePlayerSkin(skinData.PlayerID, newHash, headSkin, time.Now().Unix())
}

func savePlayerSkinsAsync(app *application, r *http.Request, stats []database.PlayerStatistics) {
	app.backgroundTask(r, func() error {
		ticker := time.NewTicker(150 * time.Millisecond)
		defer ticker.Stop()

		for _, player := range stats {
			skinData, err := app.db.GetPlayerSkinData(player.UUID)
			if err != nil {
				app.logger.Error("Getting player skin data from DB failed", "player", player.Name, "err", err)
				continue
			}

			<-ticker.C

			if skinData.SkinHash == nil && skinData.FetchedAt == nil {
				if err := saveNewPlayerSkin(app.db, skinData, player.Name); err != nil {
					app.logger.Error("Saving new player skin data failed", "player", player.Name, "err", err)
				}
				continue
			}

			if err = updateExistingPlayerSkin(app.db, skinData, player.Name); err != nil {
				app.logger.Error("Updating existing player skin data failed", "player", player.Name, "err", err)
			}

		}
		return nil
	})
}
