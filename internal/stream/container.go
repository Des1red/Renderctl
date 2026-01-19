package stream

import (
	"errors"
	"renderctl/internal/servers"
	"strings"
)

type tsContainer struct{}

func (tsContainer) Key() string { return "ts" }

func (tsContainer) MimeCandidates() []string {
	return []string{
		"video/mpeg",               // most DLNA TVs accept this
		"application/octet-stream", // very permissive fallback
		"video/mp2t",               // least compatible
	}
}

type passthroughContainer struct{}

func (p passthroughContainer) Key() string { return "passthrough" }

func (p passthroughContainer) MimeCandidates() []string {
	return []string{
		"video/mp4",
		"video/mpeg",
		"application/octet-stream",
	}
}

// Registry for containers
var containerRegistry = map[string]servers.StreamContainer{
	"ts":          tsContainer{},
	"passthrough": passthroughContainer{},
}

func GetContainer(key string) (servers.StreamContainer, error) {
	c, ok := containerRegistry[strings.ToLower(strings.TrimSpace(key))]
	if !ok {
		return nil, errors.New("unknown container: " + key)
	}
	return c, nil
}
