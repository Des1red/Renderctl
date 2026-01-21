package avtransport

import (
	"renderctl/internal/cache"
	"renderctl/internal/identity"
	"renderctl/internal/models"
	myidentity "renderctl/internal/servers/identity"
	"renderctl/internal/ssdp"
	"renderctl/internal/utils"
	"renderctl/logger"
	"strings"
	"time"
)

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
	} else {
		logger.Result(
			"SSDP discovery found %d cacheable AVTransport device(s)",
			len(found),
		)
	}
	selfUUID, _ := myidentity.FetchUUID()

	for _, tv := range found {
		// Ignore self
		if selfUUID != "" && tv.UDN == "uuid:"+selfUUID {
			logger.Info("Ignoring self SSDP MediaServer (%s)", tv.UDN)
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
