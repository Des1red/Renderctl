package internal

import (
	"net/http"
	"tvctrl/logger"
)

func ServeDirGo(cfg Config) {
	fs := http.FileServer(http.Dir(cfg.LDir))

	go func() {
		logger.Success("Go HTTP server serving: %s", cfg.LDir)
		err := http.ListenAndServe("0.0.0.0:"+cfg.ServePort, fs)
		if err != nil {
			logger.Fatal("HTTP server error: %v", err)
		}
	}()
}
