package models

import "time"

type Config struct {
	Interactive bool
	Mode        string // "auto" | "manual"
	Verbose     bool

	ProbeOnly bool

	SelectCache  int
	CacheDetails int
	AutoCache    bool
	UseCache     bool
	ForgetCache  string // "", "all", or IP
	ListCache    bool
	ShowMedia    string
	ShowMediaAll bool
	Showactions  bool

	Discover    bool
	DeepSearch  bool
	Subnet      string
	SSDPTimeout time.Duration

	TIP      string // TV IP
	TPort    string // TV SOAP port
	TPath    string // SOAP path
	TVVendor string // TV vendor

	LIP       string // local IP
	LFile     string // local file path (used only for MediaURL)
	LDir      string // directory to serve
	ServePort string // local HTTP port

	CachedConnMgrURL string
	CachedControlURL string
	ServerUp         bool
}

var DefaultConfig = Config{
	// Ssdp
	SSDPTimeout: 60 * time.Second,
	Discover:    false,
	// Cache
	Interactive:  false,
	SelectCache:  -1,
	CacheDetails: -1,
	ShowMedia:    "",
	ShowMediaAll: false,
	Showactions:  false,
	AutoCache:    false,
	UseCache:     true,

	ProbeOnly: false,
	Mode:      "auto",
	ServePort: "8000",
	LDir:      "./directory",
	Verbose:   false,
}
