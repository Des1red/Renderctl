package identity

import "net/http"

func PolishHeaders(w http.ResponseWriter) {
	// Identify as a boring DLNA appliance, not Go
	w.Header().Set("Server", "renderctl/1.0 DLNA/1.0")

	// Prevent Go from adding odd defaults
	w.Header().Del("X-Powered-By")
}
