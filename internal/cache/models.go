package cache

type Device struct {
	Vendor     string `json:"vendor"`
	ControlURL string `json:"control_url"`

	// --- enrichment
	Identity map[string]any      `json:"identity,omitempty"`
	Actions  map[string]bool     `json:"actions,omitempty"`
	Media    map[string][]string `json:"media,omitempty"`
}

type Store map[string]Device // keyed by TV IP
