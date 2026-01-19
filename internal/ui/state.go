package ui

import (
	"renderctl/internal/models"
	"time"
)

type uiState int

const (
	stateModeSelect uiState = iota
	stateConfig
	stateConfirm
	stateExit
	stateBoot
)

// UI-local transactional state
type uiContext struct {
	// original config (never mutated by UI)
	cfg *models.Config

	// working copy edited by TUI
	working models.Config

	ssdpTimeoutSec int
}

// reset working copy to defaults
func (u *uiContext) resetWorking() {
	u.working = models.DefaultConfig
	u.ssdpTimeoutSec = int(u.working.SSDPTimeout / time.Second)
}

// copy working config into real config (commit point, used later)
func (u *uiContext) commit() {
	*u.cfg = u.working
}
