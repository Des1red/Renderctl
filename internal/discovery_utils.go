package internal

import (
	"fmt"
	"time"
	"tvctrl/internal/avtransport"
	"tvctrl/internal/cache"
	"tvctrl/internal/identity"
	"tvctrl/internal/ssdp"
	"tvctrl/logger"
)

func trySSDP(cfg *Config) bool {
	logger.Notify("Running SSDP discovery scan")
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

func probeAVTransport(cfg *Config) (bool, error) {
	if cfg.TIP == "" {
		return false, fmt.Errorf("probe requires -Tip")
	}

	logger.Notify("Probing AVTransport directly : %s", cfg.TIP)

	target, err := avtransport.Probe(cfg.TIP, 8*time.Second)
	if err != nil {
		return false, err
	}

	// update cfg so playback can continue
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

	storeInCache(cfg, target)

	return true, nil
}
