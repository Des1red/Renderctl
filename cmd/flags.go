package cmd

import (
	"os"
	"renderctl/internal/models"
	"renderctl/requirements"

	"github.com/spf13/pflag"
)

func parseFlags() {
	//installation
	pflag.BoolVar(&requirements.Install, "install", false, "Run installer (build binary and optional dependencies)")
	pflag.BoolVar(&requirements.DryRun, "dry-run", false, "Show installer actions without executing")

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
	pflag.StringVar(&cfg.ShowMedia, "show-media", cfg.ShowMedia, "Show media details (audio,video,image or comma-separated)")
	pflag.BoolVar(&cfg.ShowMediaAll, "show-media-all", cfg.ShowMediaAll, "Show all media information from cached devices")
	pflag.BoolVar(&cfg.Showactions, "show-actions", cfg.Showactions, "Show supported actions from cached devices")

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

	// output
	pflag.BoolVar(&cfg.Verbose, "verbose", cfg.Verbose, "Enables verbose output")

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

func badFlagUse() (bool, string) {
	def := models.DefaultConfig

	// scan mode restrictions
	if cfg.Mode == "scan" && cfg.ProbeOnly {
		return true, "flag --probe-only is not supported in scan mode"
	}
	if cfg.Mode == "scan" && !cfg.Discover && cfg.TIP == "" && cfg.Subnet == "" {
		return true, "scan mode requires a target IP or subnet when SSDP discovery is disabled"
	}

	// cache mode conflicts
	if cfg.ListCache && cfg.CacheDetails >= 0 {
		return true, "flags --list-cache and --details-cache cannot be used together"
	}

	// cached target overrides
	if cfg.SelectCache != def.SelectCache &&
		(cfg.TIP != def.TIP ||
			cfg.TPort != def.TPort ||
			cfg.TPath != def.TPath ||
			cfg.TVVendor != def.TVVendor) {
		return true, "cannot override TV connection parameters when a cached target is selected"
	}

	// SSDP flag dependency
	if !cfg.Discover && cfg.SSDPTimeout != def.SSDPTimeout {
		return true, "flag --ssdp-timeout requires SSDP discovery to be enabled (--ssdp)"
	}

	return false, ""
}
