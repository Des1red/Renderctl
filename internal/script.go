package internal

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"
	"tvctrl/internal/avtransport"
	"tvctrl/internal/cache"
	"tvctrl/internal/identity"
	"tvctrl/internal/ssdp"
	"tvctrl/logger"
)

func RunScript(cfg Config) {
	switch cfg.Mode {
	case "manual":
		runManual(cfg)
	case "auto":
		runAuto(cfg)
	default:
		log.Fatalf("Unknown mode: %s", cfg.Mode)
	}
}

func runManual(cfg Config) {
	target := avtransport.Target{
		ControlURL: cfg.ControlURL(),
		MediaURL:   cfg.MediaURL(),
	}

	meta := avtransport.MetadataForVendor(cfg.TVVendor, target)
	avtransport.Run(target, meta)
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
	ok, err := probeAVTransport(&cfg)
	if err != nil {
		logger.Fatal("Error: %v", err)
	}

	if cfg.ProbeOnly {
		logger.Success("Probe completed (no playback).")
		os.Exit(0)
	} else {
		logger.Success("Probe completed (sending file).")
	}

	if ok {
		runWithConfig(cfg)
		return
	}
	logger.Fatal("Unable to resolve AVTransport endpoint")
}

func runWithConfig(cfg Config) {
	controlURL := cfg.ControlURL()

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

func trySSDP(cfg *Config) bool {
	devices, _ := ssdp.ListenNotify(3 * time.Second)

	if len(devices) == 0 {
		logger.Notify("No NOTIFY devices from SSDP listen, trying SSDP discover")
		devices, _ = ssdp.Discover(3 * time.Second)
	}

	if len(devices) == 0 {
		return false
	}

	tv, err := ssdp.FetchAndDetect(devices[0].Location)
	if err != nil {
		return false
	}

	if tv.IP != "" {
		cfg.TIP = tv.IP
	}
	if tv.Port != "" {
		cfg.TPort = tv.Port
	}
	if tv.ControlURL != "" {
		cfg.TPath = tv.ControlURL
	}
	if tv.Vendor != "" {
		cfg.TVVendor = tv.Vendor
	}

	caps, err := avtransport.EnrichCapabilities(
		tv.AVTransportSCPD,
		tv.ConnectionManagerCtrl,
		avtransport.Target{
			ControlURL: cfg.ControlURL(),
		},
	)

	info, err := identity.Enrich(
		cfg.BaseUrl(),
		3*time.Second,
	)

	if err == nil {
		store, _ := cache.Load()
		dev := store[cfg.TIP]

		dev.Identity = map[string]any{
			"friendly_name": info.FriendlyName,
			"manufacturer":  info.Manufacturer,
			"model_name":    info.ModelName,
			"model_number":  info.ModelNumber,
			"udn":           info.UDN,
			"presentation":  info.Presentation,
		}

		store[cfg.TIP] = dev
		_ = cache.Save(store)
	}

	if err == nil {
		store, _ := cache.Load()
		dev := store[cfg.TIP]

		// preserve existing fields
		if dev.ControlURL == "" {
			dev.ControlURL = cfg.ControlURL()
		}
		if dev.Vendor == "" {
			dev.Vendor = tv.Vendor
		}

		dev.Actions = caps.Actions
		dev.Media = caps.Media

		store[cfg.TIP] = dev
		_ = cache.Save(store)
	}

	return true
}

func confirm(msg string) bool {
	var ans string
	logger.Info("%s (y/n): ", msg)
	fmt.Scanln(&ans)
	ans = strings.ToLower(ans)
	return ans == "y" || ans == "yes"
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
	cfg._CachedControlURL = dev.ControlURL // see next fix

	return true
}

func probeAVTransport(cfg *Config) (bool, error) {
	if cfg.TIP == "" {
		return false, fmt.Errorf("probe requires -Tip")
	}

	logger.Notify("Probing AVTransport directly")

	target, err := avtransport.Probe(cfg.TIP, 8*time.Second)
	if err != nil {
		return false, err
	}

	// IMPORTANT: update cfg so playback can continue
	cfg._CachedControlURL = target.ControlURL
	info, err := identity.Enrich(
		"http://"+cfg.TIP,
		3*time.Second,
	)
	if err == nil {
		store, _ := cache.Load()
		dev := store[cfg.TIP]

		dev.Identity = map[string]any{
			"friendly_name": info.FriendlyName,
			"manufacturer":  info.Manufacturer,
			"model_name":    info.ModelName,
			"model_number":  info.ModelNumber,
			"udn":           info.UDN,
			"presentation":  info.Presentation,
		}

		// preserve known fields
		if dev.ControlURL == "" {
			dev.ControlURL = target.ControlURL
		}
		if dev.Vendor == "" {
			dev.Vendor = cfg.TVVendor
		}

		store[cfg.TIP] = dev
		_ = cache.Save(store)
	} else {
		logger.Notify("%v", err)
	}

	logger.Success("\n=== AVTransport Probe Summary ===")

	logger.Result(" IP        : %s", cfg.TIP)
	logger.Result(" ControlURL: %s", target.ControlURL)

	if confirm("Store this AVTransport endpoint in cache?") {
		store, _ := cache.Load()
		dev := store[cfg.TIP]

		if dev.ControlURL == "" {
			dev.ControlURL = target.ControlURL
		}
		if dev.Vendor == "" {
			dev.Vendor = cfg.TVVendor
		}

		store[cfg.TIP] = dev
		_ = cache.Save(store)
	}

	return true, nil
}
