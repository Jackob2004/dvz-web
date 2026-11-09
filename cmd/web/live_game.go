package main

import (
	"fmt"
	"sync"

	"github.com/google/uuid"
)

type liveGameService struct {
	mu          sync.Mutex
	subscribers map[string]chan gameInfo
	gameInfo    gameInfo
}

func newLiveGamePanel() *liveGameService {
	return &liveGameService{
		subscribers: make(map[string]chan gameInfo),
		gameInfo:    gameInfo{GamePhase: "Offline"},
	}
}

func (l *liveGameService) subscribe() (bufferedChannel <-chan gameInfo, subUUID string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	channel := make(chan gameInfo, 1)
	subUUID = uuid.NewString()
	l.subscribers[subUUID] = channel
	channel <- l.gameInfo
	bufferedChannel = channel

	return
}

func (l *liveGameService) unsubscribe(subUUID string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	delete(l.subscribers, subUUID)
}

func (l *liveGameService) updateGameInfo(info gameInfo) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.gameInfo = info
	for _, sub := range l.subscribers {
		select {
		case sub <- info:
		default:
		}
	}
}

// Some *fields are null based on the GamePhase
type gameInfo struct {
	GamePhase       string `json:"gamePhase"`
	Time            int    `json:"time"`
	PlayersToStart  *int   `json:"playersToStart"`
	Dwarves         *int   `json:"dwarves"`
	Zombies         *int   `json:"zombies"`
	Shrines         *int   `json:"shrines"`
	CurrShrine      *int   `json:"currShrine"`
	ShrineHealth    *int   `json:"shrineHealth"`
	ShrineMaxHealth *int   `json:"shrineMaxHealth"`
	ShrineShielded  *bool  `json:"shrineShielded"`
}

func (g gameInfo) unpack() gameInfoView {
	return gameInfoView{
		GamePhase:       g.GamePhase,
		Time:            g.Time,
		PlayersToStart:  deref(g.PlayersToStart),
		Dwarves:         deref(g.Dwarves),
		Zombies:         deref(g.Zombies),
		Shrines:         make([]struct{}, deref(g.Shrines)),
		CurrShrine:      deref(g.CurrShrine),
		ShrineHealth:    deref(g.ShrineHealth),
		ShrineMaxHealth: deref(g.ShrineMaxHealth),
		ShrineShielded:  deref(g.ShrineShielded),
	}
}

type gameInfoView struct {
	GamePhase       string     `json:"gamePhase"`
	Time            int        `json:"time"`
	PlayersToStart  int        `json:"playersToStart"`
	Dwarves         int        `json:"dwarves"`
	Zombies         int        `json:"zombies"`
	Shrines         []struct{} `json:"shrines"`
	CurrShrine      int        `json:"currShrine"`
	ShrineHealth    int        `json:"shrineHealth"`
	ShrineMaxHealth int        `json:"shrineMaxHealth"`
	ShrineShielded  bool       `json:"shrineShielded"`
}

const templatePrefix = "live-game:"

func (g gameInfo) templateName() (string, error) {
	var suffix string
	switch g.GamePhase {
	case "Recruiting":
		suffix = "recruiting"
	case "Preparation":
		suffix = "preparation"
	case "Attack":
		suffix = "attack"
	case "Restart":
		suffix = "restart"
	case "End":
		suffix = "end"
	case "Offline":
		suffix = "offline"
	default:
		return "", fmt.Errorf("unknown game phase: %s", g.GamePhase)
	}

	return templatePrefix + suffix, nil
}
