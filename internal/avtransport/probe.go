package avtransport

import (
	"errors"
	"fmt"
	"renderctl/internal/cache"
	"renderctl/internal/identity"
	"renderctl/internal/models"
	"renderctl/logger"
	"time"
)

var probePorts = []string{
	"9197", // Samsung AVTransport
	"7678", // Samsung AllShare
	"8187",
	"9119",
	"8080",
}

var defaultList = []string{
	"/dmr/upnp/control/AVTransport1",
	"/upnp/control/AVTransport",
	"/MediaRenderer/AVTransport/Control",
	"/AVTransport/control",
}

var bigList = []string{
	// --- Common / standard ---
	"/upnp/control/AVTransport",
	"/AVTransport/control",
	"/MediaRenderer/AVTransport/Control",
	"/dmr/upnp/control/AVTransport1",

	// --- Samsung ---
	"/smp_7_/AVTransport",
	"/smp_9_/AVTransport",
	"/upnp/control/AVTransport1",

	// --- Sony / Bravia ---
	"/sony/AVTransport",
	"/upnp/control/AVTransport/1",

	// --- LG webOS ---
	"/upnp/control/AVTransport",
	"/upnp/control/avtransport",

	// --- Generic DMR variants ---
	"/renderer/control/AVTransport",
	"/device/AVTransport/control",
	"/control/AVTransport",

	// --- Case / path quirks ---
	"/AVTransport/Control",
	"/avtransport/control",
}

func TryProbe(cfg *models.Config) bool {
	ok, err := probeAVTransport(cfg)
	if err != nil {
		logger.Error("Error: %v", err)
	}
	return ok
}

func probeEndpoint(ip string, timeout time.Duration, list bool) (*Target, error) {
	var endpoints []string
	if list {
		endpoints = bigList
	} else {
		endpoints = defaultList
	}
	deadline := time.Now().Add(timeout)

	for _, port := range probePorts {
		for _, path := range endpoints {
			if time.Now().After(deadline) {
				return nil, errors.New("AVTransport probe timed out")
			}

			controlURL := fmt.Sprintf("http://%s:%s%s", ip, port, path)

			ok := probeSOAPEndpoint(controlURL, "")
			if ok {
				return &Target{
					ControlURL: controlURL,
				}, nil
			}
		}
	}

	return nil, errors.New("no AVTransport endpoint found")
}

func probeAVTransport(cfg *models.Config) (bool, error) {
	if cfg.TIP == "" {
		return false, fmt.Errorf("probe requires -Tip")
	}

	logger.Notify("Probing AVTransport directly: %s", cfg.TIP)

	target, err := probeEndpoint(cfg.TIP, 8*time.Second, cfg.DeepSearch)
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

	logger.Done("AVTransport probe completed")

	logger.Result(" IP        : %s", cfg.TIP)
	logger.Result(" ControlURL: %s", target.ControlURL)

	return true, nil
}
