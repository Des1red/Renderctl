// internal/serverStream.go
package stream

import (
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"renderctl/internal/models"
	"renderctl/logger"
)

type ClientProfile struct {
	IP        string
	UserAgent string
	Headers   http.Header

	WantsRange bool
	DidHEAD    bool
}

var (
	profiles = make(map[string]*ClientProfile)
	mu       sync.Mutex
)

func ServeStreamGo(
	cfg *models.Config,
	stop <-chan struct{},
	streamPath string,
	mime string,
	container StreamContainer,
	source StreamSource,
) {
	cfg.ServerUp = true

	mux := http.NewServeMux()

	mux.HandleFunc(streamPath, func(w http.ResponseWriter, r *http.Request) {
		clientIP := strings.Split(r.RemoteAddr, ":")[0]

		mu.Lock()
		p, ok := profiles[clientIP]
		if !ok {
			p = &ClientProfile{
				IP:        clientIP,
				UserAgent: r.UserAgent(),
				Headers:   r.Header.Clone(),
			}
			profiles[clientIP] = p
			logger.Notify("TV detected: %s (%s)", p.IP, p.UserAgent)
		}
		mu.Unlock()

		if r.Method == http.MethodHead {
			p.DidHEAD = true
		}

		if r.Header.Get("Range") != "" {
			p.WantsRange = true
			logger.Notify("TV %s requested Range", p.IP)
		}

		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		rc, err := source.Open()
		if err != nil {
			http.Error(w, "stream source unavailable", http.StatusServiceUnavailable)
			return
		}
		defer rc.Close()

		w.Header().Set("Content-Type", mime)

		// CHANGED: dynamic Accept-Ranges
		// Range support depends on container semantics
		switch container.Key() {
		case "ts":
			// MPEG-TS / live streams must be linear
			w.Header().Set("Accept-Ranges", "none")

		default:
			// Non-TS containers MAY support Range (future)
			if p.WantsRange {
				w.Header().Set("Accept-Ranges", "bytes")
			} else {
				w.Header().Set("Accept-Ranges", "none")
			}
		}

		if r.Method == http.MethodHead {
			w.WriteHeader(http.StatusOK)
			return
		}

		_, _ = io.Copy(w, rc)
	})

	srv := &http.Server{
		Addr:              "0.0.0.0:" + cfg.ServePort,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		logger.Success(
			"Go HTTP stream server listening: %s%s (mime=%s)",
			"http://"+cfg.LIP+":"+cfg.ServePort,
			streamPath,
			mime,
		)

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("HTTP stream server error: %v", err)
		}
	}()

	go func() {
		<-stop
		logger.Notify("Shutting down stream HTTP server")
		_ = srv.Close()
	}()
}
