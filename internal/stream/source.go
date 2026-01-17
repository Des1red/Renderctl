package stream

import (
	"errors"
	"renderctl/internal/models"
)

// This is generic in structure; later you replace with screen capture sources etc.
func BuildStreamSource(cfg *models.Config) (StreamSource, error) {
	kind := ResolveStreamKind(cfg)

	switch kind {
	case StreamResolved:
		return newResolverSource(cfg.LFile), nil

	case StreamExternal:
		return urlSource{url: cfg.LFile}, nil

	case StreamFile:
		return fileSource{path: cfg.LFile}, nil
	}

	return nil, errors.New("unknown stream kind")
}
