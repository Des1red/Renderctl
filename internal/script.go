package internal

import (
	"log"
	"renderctl/internal/avtransport"
	"renderctl/internal/models"
	"renderctl/internal/stream"
	"renderctl/internal/utils"
	"renderctl/logger"
)

func runWithConfig(cfg *models.Config) {
	controlURL := cfg.CachedControlURL
	if controlURL == "" {
		controlURL = utils.ControlURL(cfg)
	}

	if controlURL == "" {
		logger.Fatal("No AVTransport ControlURL resolved (internal state error)")
		return
	}

	logger.Info("Control Url : %s", controlURL)

	target := avtransport.Target{
		ControlURL: controlURL,
		MediaURL:   utils.MediaURL(cfg),
	}

	meta := avtransport.MetadataForVendor(cfg.TVVendor, target)
	avtransport.Run(target, meta)
}

func RunScript(cfg *models.Config, stop <-chan struct{}) {
	if cfg.SelectCache != -1 {
		logger.Notify("Using explicitly selected cached device")
		runWithConfig(cfg)
		return
	}
	mode := utils.NormalizeMode(cfg.Mode)
	switch mode {
	case "stream":
		runStream(cfg, stop)
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

func runAuto(cfg *models.Config) {
	// 1) SSDP
	if avtransport.TrySSDP(cfg) {
		runWithConfig(cfg)
		return
	}

	if cfg.UseCache {
		// 2) Cache (interactive)
		if avtransport.TryCache(cfg) {
			runWithConfig(cfg)
			return
		}
	}

	// 3) Probe fallback
	ok := avtransport.TryProbe(cfg)
	if !ok {
		logger.Fatal("Unable to resolve AVTransport endpoint")
	}

	if cfg.ProbeOnly {
		logger.Success("Probe completed (no playback).")
		return
	}

	runWithConfig(cfg)
}

func runManual(cfg *models.Config) {
	target := avtransport.Target{
		ControlURL: utils.ControlURL(cfg),
		MediaURL:   utils.MediaURL(cfg),
	}

	meta := avtransport.MetadataForVendor(cfg.TVVendor, target)
	avtransport.Run(target, meta)
}

func runScan(cfg *models.Config) {
	// --- SSDP scan ---
	if cfg.Discover {
		if avtransport.TrySSDP(cfg) {
			logger.Success("Device discovered via SSDP")
		} else {
			logger.Notify("No devices discovered via SSDP")
		}
		return
	}

	// --- Subnet scan ---
	if cfg.Subnet != "" {
		avtransport.ScanSubnet(cfg)
		return
	}

	// --- Single-IP probe ---
	avtransport.TryProbe(cfg)

	logger.Success("Mode : Scan , completed")
}

func runStream(cfg *models.Config, stop <-chan struct{}) {
	// Implemented in internal/stream_mode.go (next section)
	stream.StartStreamPlay(cfg, stop)
}
