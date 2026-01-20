package cache

import "time"

/*
======== STORAGE MODELS ========
*/

// Legacy view model (still used by list + playback)
type Device struct {
	Vendor     string `json:"vendor"`
	ControlURL string `json:"control_url"`
	ConnMgrURL string `json:"conn_mgr_url,omitempty"`

	Identity map[string]any      `json:"identity,omitempty"`
	Actions  map[string]bool     `json:"actions,omitempty"`
	Media    map[string][]string `json:"media,omitempty"`
}

// New internal storage model
type Store map[string]*CachedDevice // keyed by IP

type CachedDevice struct {
	Vendor    string               `json:"vendor,omitempty"`
	Identity  map[string]any       `json:"identity,omitempty"`
	Endpoints map[string]*Endpoint `json:"endpoints"` // key = ControlURL
}

type Endpoint struct {
	ControlURL string              `json:"control_url"`
	ConnMgrURL string              `json:"conn_mgr_url,omitempty"`
	Actions    map[string]bool     `json:"actions,omitempty"`
	Media      map[string][]string `json:"media,omitempty"`
	SeenAt     time.Time           `json:"seen_at"`
}
