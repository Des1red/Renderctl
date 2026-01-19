package servers

import (
	"net/http"
	"renderctl/internal/models"
	"renderctl/internal/servers/identity"
	"renderctl/logger"
)

func InitDefaultServer(cfg models.Config, stop <-chan struct{}) {
	serverUUID, err := identity.FetchUUID()
	if err != nil {
		logger.Fatal("Failed to load server UUID: %v", err)
	}

	cfg.ServerUp = true
	fs := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		identity.PolishHeaders(w)
		http.FileServer(http.Dir(cfg.LDir)).ServeHTTP(w, r)
	})

	mux := http.NewServeMux()

	identity.RegisterHandlers(mux, serverUUID)
	mux.Handle("/", fs)

	srv := &http.Server{
		Addr:    "0.0.0.0:" + cfg.ServePort,
		Handler: mux,
	}

	go func() {
		logger.Success("HTTP server serving: %s", cfg.LDir)

		identity.AnnounceMediaServer(
			serverUUID,
			"http://"+cfg.LIP+":"+cfg.ServePort+"/device.xml",
		)

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("HTTP server error: %v", err)
		}
	}()

	go func() {
		<-stop
		logger.Notify("Shutting down HTTP server")
		identity.AnnounceMediaServerByeBye(serverUUID)
		_ = srv.Close()
	}()
}
