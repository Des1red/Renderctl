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

	logger.Status("================== SSDP DEVICE ==================")
	logger.Status("===============================================")
	logger.Status("IP        : %s", cfg.TIP)
	logger.Status("Vendor    : %s", update.Vendor)
	logger.Status("ControlURL: %s", update.ControlURL)
	logger.Status("ConnMgr   : %s", update.ConnMgrURL)

	if update.Identity != nil {
		logger.Status("Name      : %v", update.Identity["friendly_name"])
		logger.Status("Model     : %v", update.Identity["model_name"])
		logger.Status("UDN       : %v", update.Identity["udn"])
	}

	if update.Actions != nil {
		logger.Status("Actions   : %d supported", len(update.Actions))
	}

	if update.Media != nil {
		logger.Status("Media     : %d profiles", len(update.Media))
	}

	logger.Status("===============================================")

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
		logger.Error("Invalid cache index: %d", cfg.SelectCache)
	}

	cfg.TIP = ip
	cfg.TVVendor = dev.Vendor
	cfg.CachedControlURL = dev.ControlURL
	cfg.CachedConnMgrURL = dev.ConnMgrURL

	logger.Notify(
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
func orNA(v string) string {
	if v == "" {
		return "n/a"
	}
	return v
}
func handleCacheDetails(cfg models.Config) {
	index := cfg.CacheDetails

	store, err := Load()
	if err != nil {
		logger.Error("Error: %v", err)
	}

	if len(store) == 0 {
		logger.Status("Cache is empty.")
		return
	}

	// deterministic IP order
	keys := make([]string, 0, len(store))
	for ip := range store {
		keys = append(keys, ip)
	}
	sort.Strings(keys)

	if index < 0 || index >= len(keys) {
		logger.Error("Invalid cache index: %d", index)
	}

	ip := keys[index]
	cd := store[ip]

	mediaFilter := mediaFilter(cfg)

	// ---- ROOT ----
	fmt.Printf("\n%s (%s)\n", ip, orNA(cd.Vendor))
	fmt.Println("├── AVTransport")

	// ---- ENDPOINTS ----
	var urls []string
	for u := range cd.Endpoints {
		urls = append(urls, u)
	}
	sort.Strings(urls)

	for i, u := range urls {
		ep := cd.Endpoints[u]

		lastEP := i == len(urls)-1
		prefix := "│   ├──"
		child := "│   │   "
		if lastEP {
			prefix = "│   └──"
			child = "│       "
		}

		playable := len(ep.Actions) > 0

		fmt.Printf("%s %s\n", prefix, ep.ControlURL)
		fmt.Printf("%s├── playable: %v\n", child, playable)

		// ---- ACTIONS ----
		if playable && len(ep.Actions) > 0 {
			fmt.Printf("%s├── actions: %d\n", child, len(ep.Actions))

			if cfg.Showactions {
				var acts []string
				for a := range ep.Actions {
					acts = append(acts, a)
				}
				sort.Strings(acts)

				for i, a := range acts {
					lastA := i == len(acts)-1 && mediaFilter == nil
					p := "├──"
					if lastA {
						p = "└──"
					}
					fmt.Printf("%s│   %s %s\n", child, p, a)
				}
			}
		}

		// ---- MEDIA ----
		if playable && len(ep.Media) > 0 {
			fmt.Printf("%s├── media: %d types\n", child, len(ep.Media))

			typeMap := map[string][]string{}
			for mime := range ep.Media {
				group := strings.SplitN(mime, "/", 2)[0]
				typeMap[group] = append(typeMap[group], mime)
			}

			var groups []string
			for g := range typeMap {
				groups = append(groups, g)
			}
			sort.Strings(groups)

			for _, g := range groups {
				expand :=
					mediaFilter != nil &&
						(mediaFilter["*"] || mediaFilter[g])

				if !expand {
					fmt.Printf(
						"%s│   ├── %s (%d)\n",
						child,
						g,
						len(typeMap[g]),
					)
					continue
				}

				fmt.Printf(
					"%s│   ├── %s (%d)\n",
					child,
					g,
					len(typeMap[g]),
				)

				sort.Strings(typeMap[g])
				for i, m := range typeMap[g] {
					p := "├──"
					if i == len(typeMap[g])-1 {
						p = "└──"
					}
					fmt.Printf(
						"%s│   │   %s %s\n",
						child,
						p,
						m,
					)
				}
			}
		}

		// ---- SEEN ----
		fmt.Printf(
			"%s└── seen: %s\n",
			child,
			ep.SeenAt.Format("2006-01-02 15:04"),
		)
	}

	fmt.Println()
}

func mediaFilter(cfg models.Config) map[string]bool {
	if cfg.ShowMediaAll {
		return map[string]bool{"*": true}
	}
	if cfg.ShowMedia == "" {
		return nil
	}

	m := map[string]bool{}
	for _, v := range strings.Split(cfg.ShowMedia, ",") {
		v = strings.TrimSpace(v)
		if v != "" {
			m[v] = true
		}
	}
	return m
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
		handleCacheDetails(cfg)
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
		logger.Error("Error: %v", err)
	}

	if len(store) == 0 {
		logger.Status("Cache is empty.")
		return
	}

	logger.Status("\n\nCached AVTransport devices:\n\n")
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
		logger.Error("%v", err)
	}

	if len(store) == 0 {
		logger.Status("Cache is empty.")
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
