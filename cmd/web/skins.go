package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"time"

	"github.com/Jackob2004/dvz-web/assets"
	"github.com/Jackob2004/dvz-web/internal/database"
)

type SkinsService struct {
	db              *database.DB
	logger          *slog.Logger
	customClient    *http.Client
	DefaultSkin     []byte
	DefaultSkinEtag string
}

func newSkinsService(db *database.DB, logger *slog.Logger) (*SkinsService, error) {
	data, err := fs.ReadFile(assets.EmbeddedFiles, "static/img/skin_head_default.png")
	if err != nil {
		return nil, err
	}

	return &SkinsService{
		db:              db,
		logger:          logger,
		customClient:    &http.Client{Timeout: 5 * time.Second},
		DefaultSkin:     data,
		DefaultSkinEtag: "\"" + hashImage(data) + "\"",
	}, nil
}

func (s *SkinsService) fetchPlayerHeadSkin(playerName string) ([]byte, error) {
	resp, err := s.customClient.Get("https://api.mcheads.org/head/" + playerName + "/64")
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

func (s *SkinsService) saveNewPlayerSkin(skinData database.PlayerSkinData, playerName string) error {
	headSkin, err := s.fetchPlayerHeadSkin(playerName)
	if err != nil {
		return err
	}

	return s.db.UpdatePlayerSkin(skinData.PlayerID, hashImage(headSkin), headSkin, time.Now().Unix())
}

func (s *SkinsService) updateExistingPlayerSkin(skinData database.PlayerSkinData, playerName string) error {
	oneWeek := time.Hour * 167
	if time.Since(time.Unix(*skinData.FetchedAt, 0)) < oneWeek {
		return nil
	}

	headSkin, err := s.fetchPlayerHeadSkin(playerName)
	if err != nil {
		return err
	}

	newHash := hashImage(headSkin)
	if newHash == *skinData.SkinHash {
		return nil
	}

	return s.db.UpdatePlayerSkin(skinData.PlayerID, newHash, headSkin, time.Now().Unix())
}

func (s *SkinsService) savePlayerSkins(stats []database.PlayerStatistics) error {
	ticker := time.NewTicker(150 * time.Millisecond)
	defer ticker.Stop()

	for _, player := range stats {
		skinData, err := s.db.GetPlayerSkinData(player.UUID)
		if err != nil {
			s.logger.Error("Getting player skin data from DB failed", "player", player.Name, "err", err)
			continue
		}

		<-ticker.C

		if skinData.SkinHash == nil && skinData.FetchedAt == nil {
			if err := s.saveNewPlayerSkin(skinData, player.Name); err != nil {
				s.logger.Error("Saving new player skin data failed", "player", player.Name, "err", err)
			}
			continue
		}

		if err = s.updateExistingPlayerSkin(skinData, player.Name); err != nil {
			s.logger.Error("Updating existing player skin data failed", "player", player.Name, "err", err)
		}

	}
	return nil
}

func hashImage(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}
