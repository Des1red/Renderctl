package internal

import (
	"log"
	"strings"
	"tvctrl/internal/avtransport"
	"tvctrl/internal/cache"
	"tvctrl/logger"
)

func runWithConfig(cfg Config) {

	controlURL := cfg.ControlURL()
	if controlURL == "" {
		logger.Fatal("No AVTransport ControlURL resolved (internal state error)")
	} else {
		logger.Info("Control Url : %s", cfg.ControlURL())
	}

	if cfg._CachedControlURL != "" {
		controlURL = cfg._CachedControlURL
	}

	target := avtransport.Target{
		ControlURL: controlURL,
		MediaURL:   cfg.MediaURL(),
	}

	meta := avtransport.MetadataForVendor(cfg.TVVendor, target)
	avtransport.Run(target, meta)
}

func RunScript(cfg Config) {
	if cfg.SelectCache != -1 {
		logger.Notify("Using explicitly selected cached device")
		runWithConfig(cfg)
		return
	}
	mode := strings.ToLower(cfg.Mode)
	switch mode {
	case "scan":
		runScan(cfg)
	case "manual":
		runManual(cfg)
	case "auto":
		runAuto(cfg)
	default:
		log.Fatalf("Unknown mode: %s", cfg.Mode)
	}
}

func runAuto(cfg Config) {
	// 1) SSDP
	if trySSDP(&cfg) {
		runWithConfig(cfg)
		return
	}

	if cfg.UseCache {
		// 2) Cache (interactive)
		if tryCache(&cfg) {
			runWithConfig(cfg)
			return
		}
	}

	// 3) Probe fallback
	ok := tryProbe(&cfg)
	if !ok {
		logger.Fatal("Unable to resolve AVTransport endpoint")
	}

	if cfg.ProbeOnly {
		logger.Success("Probe completed (no playback).")
		return
	}

	runWithConfig(cfg)
}

func runManual(cfg Config) {
	target := avtransport.Target{
		ControlURL: cfg.ControlURL(),
		MediaURL:   cfg.MediaURL(),
	}

	meta := avtransport.MetadataForVendor(cfg.TVVendor, target)
	avtransport.Run(target, meta)
}

func runScan(cfg Config) {
	// --- SSDP scan ---
	if cfg.Discover {
		if trySSDP(&cfg) {
			logger.Success("Device discovered via SSDP")
		} else {
			logger.Notify("No devices discovered via SSDP")
		}
		return
	}

	// --- Subnet scan ---
	if cfg.Subnet != "" {
		scanSubnet(cfg)
		return
	}

	// --- Single-IP probe ---
	tryProbe(&cfg)

	logger.Success("Mode : Scan , completed")
}

func tryProbe(cfg *Config) bool {
	ok, err := probeAVTransport(cfg)
	if err != nil {
		logger.Fatal("Error: %v", err)
	}
	return ok
}

func tryCache(cfg *Config) bool {
	if cfg.TIP == "" {
		return false
	}

	store, _ := cache.Load()
	dev, ok := store[cfg.TIP]
	if !ok {
		return false
	}

	logger.Notify("\nCached device found:")
	logger.Result(" IP        : %s", cfg.TIP)
	logger.Result(" Vendor    : %s", dev.Vendor)
	logger.Result(" ControlURL: %s", dev.ControlURL)

	if !confirm("Use cached AVTransport endpoint?") {
		return false
	}

	//  IMPORTANT: do NOT touch TPath / ControlURL builder
	cfg.TVVendor = dev.Vendor

	// Store FULL URL directly
	cfg.TPath = ""
	cfg.TPort = ""
	cfg.TIP = ""

	// Inject directly into playback phase
	cfg._CachedControlURL = dev.ControlURL

	return true
}
