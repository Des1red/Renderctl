package cache

import (
	"fmt"
	"renderctl/internal/models"
	"renderctl/internal/utils"
	"renderctl/logger"
	"sort"
	"strings"
	"time"
)

/*
======== CACHE WRITE PATH ========
*/

func StoreInCache(cfg *models.Config, update Device) {
	if !cfg.UseCache || cfg.SelectCache != -1 {
		return
	}

	logger.Notify("=========== SSDP DEVICE ===========")
	logger.Notify("IP        : %s", cfg.TIP)
	logger.Notify("Vendor    : %s", update.Vendor)
	logger.Notify("ControlURL: %s", update.ControlURL)
	logger.Notify("ConnMgr   : %s", update.ConnMgrURL)

	if update.Identity != nil {
		logger.Notify("Name      : %v", update.Identity["friendly_name"])
		logger.Notify("Model     : %v", update.Identity["model_name"])
		logger.Notify("UDN       : %v", update.Identity["udn"])
	}

	if update.Actions != nil {
		logger.Notify("Actions   : %d supported", len(update.Actions))
	}

	if update.Media != nil {
		logger.Notify("Media     : %d profiles", len(update.Media))
	}

	logger.Notify("===================================")

	store, _ := Load()

	// ---- DEVICE LEVEL ----
	cd, ok := store[cfg.TIP]
	if !ok {
		cd = &CachedDevice{
			Vendor:    update.Vendor,
			Identity:  update.Identity,
			Endpoints: map[string]*Endpoint{},
		}
		store[cfg.TIP] = cd
	}

	if cd.Vendor == "" && update.Vendor != "" {
		cd.Vendor = update.Vendor
	}

	if update.Identity != nil {
		cd.Identity = update.Identity
	}

	// ---- ENDPOINT LEVEL ----
	if update.ControlURL != "" {
		ep, ok := cd.Endpoints[update.ControlURL]
		if !ok {
			ep = &Endpoint{
				ControlURL: update.ControlURL,
				SeenAt:     time.Now(),
			}
			cd.Endpoints[update.ControlURL] = ep
		}

		ep.SeenAt = time.Now()

		if update.ConnMgrURL != "" {
			ep.ConnMgrURL = update.ConnMgrURL
		}
		if update.Actions != nil {
			ep.Actions = update.Actions
		}
		if update.Media != nil {
			ep.Media = update.Media
		}
	}

	_ = Save(store)
}

/*
======== LEGACY READ PATH ========
*/

func LoadCachedTV(cfg *models.Config) {
	ip, dev, ok := selectFromCache(cfg.SelectCache)
	if !ok {
		logger.Fatal("Invalid cache index: %d", cfg.SelectCache)
	}

	cfg.TIP = ip
	cfg.TVVendor = dev.Vendor
	cfg.CachedControlURL = dev.ControlURL
	cfg.CachedConnMgrURL = dev.ConnMgrURL

	logger.Success(
		"Using cached device [%d]: %s",
		cfg.SelectCache,
		dev.ControlURL,
	)
}

func selectFromCache(index int) (string, Device, bool) {
	store, _ := Load()
	keys := sortedCache(store)

	if index < 0 || index >= len(keys) {
		return "", Device{}, false
	}

	ip := keys[index]
	cd := store[ip]

	// derive primary endpoint deterministically
	var urls []string
	for u, ep := range cd.Endpoints {
		if len(ep.Actions) > 0 {
			urls = append(urls, u)
		}
	}
	sort.Strings(urls)

	if len(urls) == 0 {
		return "", Device{}, false
	}

	primary := cd.Endpoints[urls[0]]

	return ip, Device{
		Vendor:     cd.Vendor,
		ControlURL: pick(primary, func(e *Endpoint) string { return e.ControlURL }),
		ConnMgrURL: pick(primary, func(e *Endpoint) string { return e.ConnMgrURL }),
		Identity:   cd.Identity,
		Actions:    pick(primary, func(e *Endpoint) map[string]bool { return e.Actions }),
		Media:      pick(primary, func(e *Endpoint) map[string][]string { return e.Media }),
	}, true
}

func pick[T any](ep *Endpoint, f func(*Endpoint) T) T {
	var zero T
	if ep == nil {
		return zero
	}
	return f(ep)
}

func handleCacheDetails(index int) {
	store, err := Load()
	if err != nil {
		logger.Fatal("Error: %v", err)
	}

	if len(store) == 0 {
		logger.Notify("Cache is empty.")
		return
	}

	// sorted IPs (same ordering everywhere)
	keys := make([]string, 0, len(store))
	for ip := range store {
		keys = append(keys, ip)
	}
	sort.Strings(keys)

	if index < 0 || index >= len(keys) {
		logger.Fatal("Invalid cache index: %d", index)
	}

	ip := keys[index]
	cd := store[ip]

	logger.Info("\n===== CACHE DETAILS =====")
	logger.Result("Index   : %d", index)
	logger.Result("IP      : %s", ip)
	logger.Result("Vendor  : %s", col(cd.Vendor, 20))

	// ---- Identity ----
	if cd.Identity != nil {
		logger.Info("\nIdentity:")
		for _, k := range []string{
			"friendly_name",
			"manufacturer",
			"model_name",
			"model_number",
			"udn",
			"presentation",
		} {
			if v, ok := cd.Identity[k]; ok {
				logger.Result(" %-14s: %v", k, v)
			}
		}
	}

	// ---- Endpoints ----
	if len(cd.Endpoints) == 0 {
		logger.Notify("\nNo AVTransport endpoints stored.")
		return
	}

	// deterministic endpoint order
	var urls []string
	for u := range cd.Endpoints {
		urls = append(urls, u)
	}
	sort.Strings(urls)

	logger.Info("\nAVTransport Endpoints:")

	for i, u := range urls {

		ep := cd.Endpoints[u]
		playable := len(ep.Actions) > 0

		logger.Info("\n [%d]", i)
		logger.Result(" Playable   : %v", playable)
		logger.Result(" ControlURL : %s", ep.ControlURL)
		logger.Result(" ConnMgrURL : %s", col(ep.ConnMgrURL, 20))
		logger.Result(" SeenAt     : %s", ep.SeenAt.Format(time.RFC3339))

		if len(ep.Actions) > 0 {
			logger.Info(" Actions:")
			var acts []string
			for a := range ep.Actions {
				acts = append(acts, a)
			}
			sort.Strings(acts)
			for _, a := range acts {
				logger.Result("  - %s", a)
			}
		}

		if len(ep.Media) > 0 {
			logger.Info(" Media:")
			var keys []string
			for k := range ep.Media {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				logger.Result("  - %s: %v", k, ep.Media[k])
			}
		}
	}

	logger.Info("\n=========================\n")
}

/*
======== CACHE COMMANDS ========
*/

func sortedCache(store Store) []string {
	keys := make([]string, 0, len(store))
	for ip := range store {
		keys = append(keys, ip)
	}
	sort.Strings(keys)
	return keys
}

func HandleCacheCommands(cfg models.Config) bool {
	if cfg.ListCache {
		handleListCache()
		return true
	}

	if cfg.CacheDetails >= 0 {
		handleCacheDetails(cfg.CacheDetails)
		return true
	}

	if cfg.ForgetCache != "" {
		handleForgetCache(cfg)
		return true
	}

	return false
}

func handleListCache() {
	store, err := Load()
	if err != nil {
		logger.Fatal("Error: %v", err)
	}

	if len(store) == 0 {
		logger.Info("Cache is empty.")
		return
	}

	logger.Info("\n\nCached AVTransport devices:\n\n")
	fmt.Printf(
		" %-3s %-15s %-12s %-60s %-60s\n",
		"#", "IP", "Vendor", "ControlURL", "ConnMgrURL",
	)
	fmt.Println(strings.Repeat("-", 160))

	keys := sortedCache(store)

	for i, ip := range keys {
		cd := store[ip]

		var urls []string
		for u, ep := range cd.Endpoints {
			if len(ep.Actions) > 0 {
				urls = append(urls, u)
			}
		}

		sort.Strings(urls)

		var ep *Endpoint
		if len(urls) > 0 {
			ep = cd.Endpoints[urls[0]]
		}

		fmt.Printf(
			"[%d] %-15s %-12s %-60s %-60s\n",
			i,
			ip,
			col(cd.Vendor, 12),
			col(pick(ep, func(e *Endpoint) string { return e.ControlURL }), 60),
			col(pick(ep, func(e *Endpoint) string { return e.ConnMgrURL }), 60),
		)
	}
}

func handleForgetCache(cfg models.Config) {
	store, err := Load()
	if err != nil {
		logger.Fatal("%v", err)
	}

	if len(store) == 0 {
		logger.Notify("Cache is empty.")
		return
	}

	switch cfg.ForgetCache {

	case "all":
		if !utils.Confirm("Delete ALL cached devices?") {
			return
		}
		_ = Save(Store{})
		logger.Success("Cache cleared.")
		return

	default:
		if _, ok := store[cfg.ForgetCache]; !ok {
			logger.Notify("IP not found in cache.")
			return
		}

		if !utils.Confirm("Delete cached entry for " + cfg.ForgetCache + "?") {
			return
		}

		delete(store, cfg.ForgetCache)
		_ = Save(store)
		logger.Success("Deleted %s", cfg.ForgetCache)
	}
}

func col(v string, w int) string {
	if v == "" {
		v = "n/a"
	}
	if len(v) > w {
		return v[:w-3] + "..."
	}
	return fmt.Sprintf("%-*s", w, v)
}
