package avtransport

import (
	"fmt"
	"renderctl/internal/cache"
	"renderctl/internal/identity"
	"renderctl/internal/models"
	"renderctl/internal/ssdp"
	"renderctl/internal/utils"
	"renderctl/logger"
	"time"
)

func TryProbe(cfg *models.Config) bool {
	ok, err := probeAVTransport(cfg)
	if err != nil {
		logger.Fatal("Error: %v", err)
	}
	return ok
}

func TryCache(cfg *models.Config) bool {
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

	if !utils.Confirm("Use cached AVTransport endpoint?") {
		return false
	}

	//  IMPORTANT: do NOT touch TPath / ControlURL builder
	cfg.TVVendor = dev.Vendor

	// Store FULL URL directly
	cfg.TPath = ""
	cfg.TPort = ""
	cfg.TIP = ""

	// Inject directly into playback phase
	cfg.CachedControlURL = dev.ControlURL
	cfg.CachedConnMgrURL = dev.ConnMgrURL

	return true
}

func TrySSDP(cfg *models.Config) bool {
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
	if tv.ConnectionManagerCtrl != "" {
		cfg.CachedConnMgrURL = tv.ConnectionManagerCtrl
	}

	caps, err := EnrichCapabilities(
		tv.AVTransportSCPD,
		tv.ConnectionManagerCtrl,
		Target{
			ControlURL: utils.ControlURL(cfg),
		},
	)

	info, err := identity.Enrich(
		utils.BaseUrl(cfg),
		3*time.Second,
	)

	update := cache.Device{
		ControlURL: utils.ControlURL(cfg),
		Vendor:     tv.Vendor,
		ConnMgrURL: tv.ConnectionManagerCtrl,
	}

	if err == nil {
		update.Identity = map[string]any{
			"friendly_name": info.FriendlyName,
			"manufacturer":  info.Manufacturer,
			"model_name":    info.ModelName,
			"model_number":  info.ModelNumber,
			"udn":           info.UDN,
			"presentation":  info.Presentation,
		}
	}

	if err == nil && caps != nil {
		update.Actions = caps.Actions
		update.Media = caps.Media
	} else {
		logger.Notify("Capability enrichment failed: %v", err)
	}

	cache.StoreInCache(cfg, update)

	return true
}

func probeAVTransport(cfg *models.Config) (bool, error) {
	if cfg.TIP == "" {
		return false, fmt.Errorf("probe requires -Tip")
	}

	logger.Notify("Probing AVTransport directly : %s", cfg.TIP)

	target, err := Probe(cfg.TIP, 8*time.Second, cfg.DeepSearch)
	if err != nil {
		return false, err
	}

	observedActions := ValidateActions(*target)

	// update cfg so playback can continue
	cfg.CachedControlURL = target.ControlURL
	info, err := identity.Enrich(
		"http://"+cfg.TIP,
		3*time.Second,
	)
	update := cache.Device{
		ControlURL: target.ControlURL,
		Vendor:     cfg.TVVendor,
	}

	if err == nil {
		update.Identity = map[string]any{
			"friendly_name": info.FriendlyName,
			"manufacturer":  info.Manufacturer,
			"model_name":    info.ModelName,
			"model_number":  info.ModelNumber,
			"udn":           info.UDN,
			"presentation":  info.Presentation,
		}
	} else {
		logger.Notify("%v", err)
	}

	if len(observedActions) > 0 {
		update.Actions = observedActions
	}

	cache.StoreInCache(cfg, update)

	logger.Success("\n=== AVTransport Probe Summary ===")

	logger.Result(" IP        : %s", cfg.TIP)
	logger.Result(" ControlURL: %s", target.ControlURL)

	return true, nil
}
