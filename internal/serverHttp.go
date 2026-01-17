package internal

import (
	"net/http"
	"renderctl/internal/models"
	"renderctl/logger"
)

func ServeDirGo(cfg models.Config, stop <-chan struct{}) {
	cfg.ServerUp = true
	fs := http.FileServer(http.Dir(cfg.LDir))

	srv := &http.Server{
		Addr:    "0.0.0.0:" + cfg.ServePort,
		Handler: fs,
	}

	go func() {
		logger.Success("Go HTTP server serving: %s", cfg.LDir)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("HTTP server error: %v", err)
		}
	}()

	go func() {
		<-stop
		logger.Notify("Shutting down HTTP server")
		_ = srv.Close()
	}()
}
