package avtransport

import (
	"fmt"
	"renderctl/internal/cache"
	"renderctl/internal/identity"
	"renderctl/internal/models"
	myidentity "renderctl/internal/servers/identity"
	"renderctl/internal/ssdp"
	"renderctl/internal/utils"
	"renderctl/logger"
	"sort"
	"strings"
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
	logger.Result(" IP        : %s", cfg.TIP)
	logger.Result(" Vendor    : %s", cd.Vendor)
	logger.Result(" ControlURL: %s", ep.ControlURL)

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

func TrySSDP(cfg *models.Config) bool {
	logger.Notify("Running SSDP discovery scan")

	devices, _ := ssdp.ListenNotify(cfg.SSDPTimeout, cfg.LIP)

	// filter + detect
	found := make([]*ssdp.DetectedTV, 0)
	for _, d := range devices {
		if d.Location == "" {
			continue
		}
		if !ssdp.LooksLikeTV(d) {
			continue
		}
		if strings.Contains(d.Location, "nservice") {
			continue
		}
		t, err := ssdp.FetchAndDetect(d.Location)
		if err != nil {
			continue
		}
		found = append(found, t)
	}

	// fallback even if NOTIFY had packets, but none were TVs
	if len(found) == 0 {
		logger.Notify("No TV-like NOTIFY devices, trying SSDP active discovery")
		devices, _ = ssdp.Discover(3 * time.Second)

		for _, d := range devices {
			if d.Location == "" {
				continue
			}
			if !ssdp.LooksLikeTV(d) {
				logger.Info("SSDP rejected: USN=%s SERVER=%s", d.USN, d.Server)
				continue
			}
			logger.Info(
				"SSDP candidate accepted: USN=%s SERVER=%s LOCATION=%s",
				d.USN,
				d.Server,
				d.Location,
			)
			if strings.Contains(d.Location, "nservice") {
				continue
			}
			t, err := ssdp.FetchAndDetect(d.Location)
			if err != nil {
				continue
			}
			found = append(found, t)
		}
	}

	if len(found) == 0 {
		logger.Notify("SSDP yielded no cacheable AVTransport targets")
		return false
	}
	selfUUID, _ := myidentity.FetchUUID()

	for _, tv := range found {
		// Ignore self
		if selfUUID != "" && tv.UDN == "uuid:"+selfUUID {
			logger.Notify("Ignoring self SSDP MediaServer (%s)", tv.UDN)
			continue
		}

		// ---- LOCAL COPY (do NOT mutate global cfg) ----
		local := *cfg

		if tv.IP != "" {
			local.TIP = tv.IP
		}
		if tv.Port != "" {
			local.TPort = tv.Port
		}
		if tv.ControlURL != "" {
			local.TPath = tv.ControlURL
		}
		if tv.Vendor != "" {
			local.TVVendor = tv.Vendor
		}
		if tv.ConnectionManagerCtrl != "" {
			local.CachedConnMgrURL = tv.ConnectionManagerCtrl
		}

		caps, err := EnrichCapabilities(
			tv.AVTransportSCPD,
			tv.ConnectionManagerCtrl,
			Target{
				ControlURL: utils.ControlURL(&local),
			},
		)

		info, infoErr := identity.Enrich(
			utils.BaseUrl(&local),
			3*time.Second,
		)

		update := cache.Device{
			ControlURL: utils.ControlURL(&local),
			Vendor:     tv.Vendor,
			ConnMgrURL: tv.ConnectionManagerCtrl,
		}

		if infoErr == nil {
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
		} else if err != nil {
			logger.Notify("Capability enrichment failed: %v", err)
		}

		cache.StoreInCache(&local, update)
	}

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
