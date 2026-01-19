package stream

import "renderctl/internal/servers"

func selectMime(
	container servers.StreamContainer,
	supported map[string][]string,
) string {
	// If TV returned nothing → safe default for streaming
	if len(supported) == 0 {
		return "video/mpeg"
	}

	for _, cand := range container.MimeCandidates() {
		if _, ok := supported[cand]; ok {
			return cand
		}
	}

	// Nothing matched → conservative fallback
	return "video/mpeg"
}
