package stream

import (
	"errors"
	"renderctl/internal/avtransport"
	"renderctl/internal/models"
	"renderctl/internal/servers"
	"strings"
)

type StreamPlan struct {
	StreamPath string
	Mime       string
	Container  servers.StreamContainer
	Source     servers.StreamSource
}

func ResolveStreamPlan(cfg *models.Config) (*StreamPlan, error) {
	// Decide container
	kind := ResolveStreamKind(cfg)

	containerKey := "ts"
	if kind == StreamExternal {
		containerKey = "passthrough"
	}

	container, err := GetContainer(containerKey)
	if err != nil {
		return nil, err
	}

	// Decide source
	src, err := BuildStreamSource(cfg)
	if err != nil {
		return nil, err
	}

	// Resolve AVTransport + protocol info (ONCE)
	if cfg.UseCache {
		if avtransport.TryCache(cfg) {
			goto resolved
		}
	}

	if !avtransport.TryProbe(cfg) {
		return nil, errors.New("unable to resolve AVTransport")
	}

resolved:
	var media map[string][]string
	if cfg.CachedConnMgrURL != "" {
		media, _ = avtransport.FetchMediaProtocols(cfg.CachedConnMgrURL)
	}
	if media == nil {
		media = map[string][]string{}
	}

	mime := selectMime(container, media)

	return &StreamPlan{
		StreamPath: "/stream",
		Mime:       mime,
		Container:  container,
		Source:     src,
	}, nil
}

type StreamKind int

const (
	StreamFile StreamKind = iota
	StreamExternal
	StreamResolved // needs yt-dlp (youtube/vimeo/etc)
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

// minimal resolver detection (expand later)
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
