package internal

type Config struct {
	Mode        string // "auto" | "manual"
	ProbeOnly   bool
	UseCache    bool
	ForgetCache string // "", "all", or IP
	ListCache   bool

	CachedTv string
	TIP      string // TV IP
	TPort    string // TV SOAP port
	TPath    string // SOAP path
	TVVendor string // TV vendor

	LIP       string // local IP
	LFile     string // local file path (used only for MediaURL)
	LDir      string // directory to serve
	ServePort string // local HTTP port

	_CachedControlURL string
}

var DefaultConfig = Config{
	UseCache:  false,
	ProbeOnly: false,
	Mode:      "auto",
	ServePort: "8000",
	LDir:      "./directory",
}
