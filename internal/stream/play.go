// internal/stream_play.go
package stream

import (
	"renderctl/internal/avtransport"
	"renderctl/internal/models"
	"renderctl/internal/utils"
	"strings"
)

type StreamKind int

const (
	StreamFile StreamKind = iota
	StreamExternal
	StreamResolved // NEW: needs yt-dlp (youtube/vimeo/etc)
)

func ResolveStreamKind(cfg *models.Config) StreamKind {
	lf := strings.TrimSpace(cfg.LFile)

	if looksLikeResolvableURL(lf) {
		return StreamResolved
	}

	if strings.HasPrefix(lf, "http://") || strings.HasPrefix(lf, "https://") {
		return StreamExternal
	}

	return StreamFile
}

// Step-1: minimal resolver detection (expand later)
func looksLikeResolvableURL(u string) bool {
	u = strings.ToLower(strings.TrimSpace(u))
	if !strings.HasPrefix(u, "http") {
		return false
	}

	return strings.Contains(u, "youtube.com") ||
		strings.Contains(u, "youtu.be") ||
		strings.Contains(u, "vimeo.com") ||
		strings.Contains(u, "twitch.tv")
}

func BuildStreamURL(cfg *models.Config, streamPath string) string {
	p := strings.TrimPrefix(streamPath, "/")
	return "http://" + cfg.LIP + ":" + cfg.ServePort + "/" + p
}

func resolveAndPlayStream(cfg *models.Config, streamPath string) {
	playStreamResolved(cfg, streamPath)
}

func playStreamResolved(cfg *models.Config, streamPath string) {
	controlURL := utils.ControlURL(cfg)
	if cfg.CachedControlURL != "" {
		controlURL = cfg.CachedControlURL
	}

	target := avtransport.Target{
		ControlURL: controlURL,
		MediaURL:   BuildStreamURL(cfg, streamPath),
	}

	meta := ""
	avtransport.Run(target, meta)
}
