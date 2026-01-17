package stream

import (
	"time"

	"renderctl/internal/avtransport"
	"renderctl/internal/models"
	"renderctl/internal/utils"
	"renderctl/logger"
)

func StartStreamPlay(cfg *models.Config, stop <-chan struct{}) {
	start(cfg, stop)
}

func start(cfg *models.Config, stop <-chan struct{}) {
	// 1) Resolve IP used for MediaURL
	cfg.LIP = utils.LocalIP(cfg.LIP)

	// 2) Decide container
	kind := ResolveStreamKind(cfg)

	containerKey := "ts"
	switch kind {
	case StreamExternal:
		containerKey = "passthrough"
	}

	container, err := GetContainer(containerKey)
	if err != nil {
		logger.Fatal("Stream container error: %v", err)
		return
	}

	// 3) Decide source (phase-1: from -Lf if provided; later: screen / remote / etc)
	src, err := BuildStreamSource(cfg)
	if err != nil {
		logger.Fatal("Stream source error: %v", err)
		return
	}

	// 4) Start stream HTTP server (/stream)
	streamPath := "/stream"
	// Resolve AVTransport ONCE
	if cfg.UseCache {
		if avtransport.TryCache(cfg) {
			goto resolved
		}
	}

	if !avtransport.TryProbe(cfg) {
		logger.Fatal("Unable to resolve AVTransport endpoint")
	}

resolved:

	controlURL := cfg.CachedControlURL
	if controlURL == "" {
		controlURL = utils.ControlURL(cfg)
	}

	if controlURL == "" {
		logger.Fatal("ControlURL not resolved")
	}

	// Fetch protocol info from ConnectionManager (if available)
	var media map[string][]string
	if cfg.CachedConnMgrURL != "" {
		media, err = avtransport.FetchMediaProtocols(cfg.CachedConnMgrURL)
	}
	if media == nil {
		logger.Notify("ProtocolInfo unavailable, using fallback MIME")
		media = map[string][]string{}
	}

	// choose MIME
	mime := selectMime(container, media)
	logger.Notify("Using stream MIME: %s", mime)
	ServeStreamGo(cfg, stop, streamPath, mime, container, src)
	for !cfg.ServerUp {
		time.Sleep(100 * time.Millisecond)
	}
	resolveAndPlayStream(cfg, streamPath)

}
