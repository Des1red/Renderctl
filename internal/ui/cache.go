package ui

import (
	"fmt"
	"net/url"
	"renderctl/internal/cache"
	"sort"
)

func openCachePopup(ctx *uiContext, index int, state *uiState) {
	ip, dev, ok := cacheSelect(index)

	if !ok {
		clearCachedSelection(ctx)
		ctx.popup = &popupState{
			kind:    popupError,
			title:   "Cache error",
			message: "Cached device not found",
		}
		*state = statePopup
		return
	}

	ctx.popup = &popupState{
		kind:     popupConfirmCache,
		title:    "Use cached device?",
		message:  formatCachedDevice(ip, dev),
		index:    index,
		ip:       ip,
		device:   dev,
		selected: 0,
	}

	*state = statePopup
}

func applyCachedDevice(ctx *uiContext, ip string, dev cache.Device) {
	ctx.working.TIP = ip
	ctx.working.TVVendor = dev.Vendor
	ctx.working.CachedControlURL = dev.ControlURL
	ctx.working.CachedConnMgrURL = dev.ConnMgrURL

	// ---- UI-only derivation from ControlURL ----
	if dev.ControlURL != "" {
		if u, err := url.Parse(dev.ControlURL); err == nil {
			if u.Port() != "" {
				ctx.working.TPort = u.Port()
			}
			ctx.working.TPath = u.Path
		}
	}
}

func cacheSelect(index int) (string, cache.Device, bool) {
	store, err := cache.Load()
	if err != nil {
		return "", cache.Device{}, false
	}

	keys := make([]string, 0, len(store))
	for ip := range store {
		keys = append(keys, ip)
	}
	sort.Strings(keys)

	if index < 0 || index >= len(keys) {
		return "", cache.Device{}, false
	}

	ip := keys[index]
	cd := store[ip]

	if cd == nil || len(cd.Endpoints) == 0 {
		return "", cache.Device{}, false
	}

	// pick primary endpoint deterministically
	var urls []string
	for u, ep := range cd.Endpoints {
		if len(ep.Actions) > 0 {
			urls = append(urls, u)
		}
	}
	sort.Strings(urls)

	ep := cd.Endpoints[urls[0]]

	return ip, cache.Device{
		Vendor:     cd.Vendor,
		ControlURL: ep.ControlURL,
		ConnMgrURL: ep.ConnMgrURL,
		Identity:   cd.Identity,
		Actions:    ep.Actions,
		Media:      ep.Media,
	}, true
}

func clearCachedSelection(ctx *uiContext) {
	ctx.working.SelectCache = -1
	ctx.working.TIP = ""
	ctx.working.TVVendor = ""
	ctx.working.CachedControlURL = ""
	ctx.working.CachedConnMgrURL = ""
}

func formatCachedDevice(ip string, dev cache.Device) string {
	return fmt.Sprintf(
		"\nIP: %s\n"+
			"\nVendor: %s\n"+
			"\nControl URL: %s\n"+
			"\nConnMgr URL: %s",
		ip,
		orNA(dev.Vendor),
		orNA(dev.ControlURL),
		orNA(dev.ConnMgrURL),
	)
}

func orNA(v string) string {
	if v == "" {
		return "n/a"
	}
	return v
}
