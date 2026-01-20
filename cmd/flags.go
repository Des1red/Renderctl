package cmd

import (
	"os"

	"github.com/spf13/pflag"
)

func parseFlags() {
	//tui startup
	pflag.BoolVar(&cfg.Interactive, "tui", cfg.Interactive, "Start program as TUI")

	pflag.BoolVar(&cfg.ProbeOnly, "probe-only", cfg.ProbeOnly, "Probe AVTransport only when using mode: auto")
	pflag.StringVar(&cfg.Mode, "mode", cfg.Mode, "Execution mode (auto/manual/stream/scan)")

	// cache
	pflag.BoolVar(&cfg.AutoCache, "auto-cache", cfg.AutoCache, "Skip cache save confirmation")
	pflag.BoolVar(&noCache, "no-cache", false, "Disable cache usage")
	pflag.BoolVar(&cfg.ListCache, "list-cache", cfg.ListCache, "List cached AVTransport devices")
	pflag.StringVar(&cfg.ForgetCache, "forget-cache", cfg.ForgetCache, "Forget cache (interactive | IP | all)")
	pflag.IntVar(&cfg.SelectCache, "select-cache", -1, "Select cached device by index")
	pflag.IntVar(&cfg.CacheDetails, "details-cache", -1, "List cached device with details")
	// scan
	pflag.StringVar(&cfg.Subnet, "subnet", cfg.Subnet, "Subnet to scan (e.g. 192.168.1.0/24)")
	pflag.BoolVar(&cfg.DeepSearch, "deep-search", cfg.DeepSearch, "Use a bigger list when probing for device endpoints")
	pflag.BoolVar(&cfg.Discover, "ssdp", cfg.Discover, "Enable SSDP discovery")
	pflag.DurationVar(
		&cfg.SSDPTimeout,
		"ssdp-timeout",
		cfg.SSDPTimeout,
		"SSDP discovery timeout (e.g. 30s, 2m)",
	)

	// tv
	pflag.StringVar(&cfg.TIP, "Tip", cfg.TIP, "TV IP address")
	pflag.StringVar(&cfg.TPort, "Tport", cfg.TPort, "TV SOAP port")
	pflag.StringVar(&cfg.TPath, "Tpath", cfg.TPath, "TV SOAP control path")
	pflag.StringVar(&cfg.TVVendor, "vendor", cfg.TVVendor, "TV vendor")

	// media
	pflag.StringVar(&cfg.LFile, "Lf", cfg.LFile, "Local media file")
	pflag.StringVar(&cfg.LIP, "Lip", cfg.LIP, "Local IP for serving media")
	pflag.StringVar(&cfg.LDir, "Ldir", cfg.LDir, "Local directory to serve")
	pflag.StringVar(&cfg.ServePort, "LPort", cfg.ServePort, "Local port to serve")

	// meta
	version := pflag.BoolP("version", "V", false, "Show version")
	help := pflag.BoolP("help", "h", false, "Show help")

	pflag.Parse()

	if *help {
		printHelp()
		os.Exit(0)
	}
	if *version {
		printVersionAndExit()
	}
}
