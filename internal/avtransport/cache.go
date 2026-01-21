package avtransport

import (
	"renderctl/internal/cache"
	"renderctl/internal/models"
	"renderctl/internal/utils"
	"renderctl/logger"
	"sort"
)

func TryProbe(cfg *models.Config) bool {
	ok, err := probeAVTransport(cfg)
	if err != nil {
		logger.Error("Error: %v", err)
	}
	return ok
}

func TryCache(cfg *models.Config) bool {
	if cfg.TIP == "" {
		return false
	}

	store, _ := cache.Load()
	cd, ok := store[cfg.TIP]
	if !ok || len(cd.Endpoints) == 0 {
		return false
	}

	// pick primary endpoint deterministically
	var urls []string
	for u, ep := range cd.Endpoints {
		if len(ep.Actions) > 0 {
			urls = append(urls, u)
		}
	}
	sort.Strings(urls)

	if len(urls) == 0 {
		return false
	}

	ep := cd.Endpoints[urls[0]]

	logger.Notify("\nCached device found:")
	logger.Status(" IP        : %s", cfg.TIP)
	logger.Status(" Vendor    : %s", cd.Vendor)
	logger.Status(" ControlURL: %s", ep.ControlURL)

	if !utils.Confirm("Use cached AVTransport endpoint?") {
		return false
	}

	// IMPORTANT: do NOT touch TPath / ControlURL builder
	cfg.TVVendor = cd.Vendor

	// Store FULL URL directly
	cfg.TPath = ""
	cfg.TPort = ""
	cfg.TIP = ""

	// Inject directly into playback phase
	cfg.CachedControlURL = ep.ControlURL
	cfg.CachedConnMgrURL = ep.ConnMgrURL

	return true
}
