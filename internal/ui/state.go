package ui

import (
	"renderctl/internal/cache"
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
	statePopup
)

// UI-local transactional state
type uiContext struct {
	bootSkip   chan struct{}
	bootDoneCh chan struct{}
	bootDone   bool

	// original config (never mutated by UI)
	cfg *models.Config

	// working copy edited by TUI
	working models.Config

	ssdpTimeoutSec int

	popup *popupState
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

type popupType int

const (
	popupConfirmCache popupType = iota
	popupError
)

type popupState struct {
	kind     popupType
	title    string
	message  string
	index    int
	ip       string
	device   cache.Device
	selected int // 0 = yes/ok, 1 = no
}
