package models

import "time"

type Config struct {
	Interactive bool
	Mode        string // "auto" | "manual"
	ProbeOnly   bool
	Discover    bool

	SelectCache int
	AutoCache   bool
	UseCache    bool
	ForgetCache string // "", "all", or IP
	ListCache   bool

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
	SSDPTimeout: 60 * time.Second,
	Interactive: false,
	SelectCache: -1,
	AutoCache:   false,
	UseCache:    true,
	ProbeOnly:   false,
	Discover:    false,
	Mode:        "auto",
	ServePort:   "8000",
	LDir:        "./directory",
}
