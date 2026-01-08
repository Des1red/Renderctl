package internal

import (
	"fmt"
	"strings"
	"tvctrl/internal/cache"
	"tvctrl/logger"
)

func HandleCacheCommands(cfg Config) bool {
	if cfg.ListCache {
		handleListCache()
		return true
	}

	if cfg.ForgetCache != "" {
		handleForgetCache(cfg)
		return true
	}

	return false
}

func handleForgetCache(cfg Config) {
	store, err := cache.Load()
	if err != nil {
		logger.Fatal("%v", err)
	}

	if len(store) == 0 {
		logger.Notify("Cache is empty.")
		return
	}

	switch cfg.ForgetCache {

	case "all":
		if !confirm("Delete ALL cached devices?") {
			return
		}
		_ = cache.Save(cache.Store{})
		logger.Success("Cache cleared.")
		return

	case "interactive":
		logger.Info("\nCached devices:")
		for ip, dev := range store {
			fmt.Printf(" %s → %s\n", ip, dev.ControlURL)
		}

		logger.Notify("\nEnter IP to delete: ")
		var ip string
		fmt.Scanln(&ip)

		if _, ok := store[ip]; !ok {
			logger.Notify("IP not found in cache.")
			return
		}

		if !confirm("Delete this entry?") {
			return
		}

		delete(store, ip)
		_ = cache.Save(store)
		logger.Success("Deleted %s", ip)
		return

	default:
		if _, ok := store[cfg.ForgetCache]; !ok {
			logger.Notify("IP not found in cache.")
			return
		}

		if !confirm("Delete cached entry for " + cfg.ForgetCache + "?") {
			return
		}

		delete(store, cfg.ForgetCache)
		_ = cache.Save(store)
		logger.Success("Deleted %s", cfg.ForgetCache)
	}
}

func handleListCache() {
	store, err := cache.Load()
	if err != nil {
		logger.Fatal("Error: %v", err)
	}

	if len(store) == 0 {
		logger.Info("Cache is empty.")
		return
	}

	logger.Info("\n\nCached AVTransport devices:\n\n")
	fmt.Printf(" %-15s %-10s %s\n", "IP", "Vendor", "ControlURL")
	fmt.Println(strings.Repeat("-", 60))

	for ip, dev := range store {
		fmt.Printf(" %-15s %-10s %s\n", ip, dev.Vendor, dev.ControlURL)
	}
}
