package cmd

import (
	"os"
	"os/signal"
	"time"

	"renderctl/internal"
	"renderctl/internal/cache"
	"renderctl/internal/models"
	"renderctl/internal/ui"
	"renderctl/internal/utils"
	"renderctl/logger"
)

var cfg = models.DefaultConfig
var noCache bool

func Execute() {
	parseFlags()

	// FLAG INVERSION
	cfg.UseCache = !noCache

	// TUI mode
	if cfg.Interactive {
		ui.Run(&cfg)
	}

	// Cache commands exit early
	if cache.HandleCacheCommands(cfg) {
		os.Exit(0)
	}

	if cfg.SelectCache >= 0 {
		cache.LoadCachedTV(&cfg)
	}

	stop := make(chan struct{})
	serverRunning := false

	// ---- PRE-RUN LOGIC ----
	mode := utils.NormalizeMode(cfg.Mode)
	if mode != "scan" && !cfg.ProbeOnly {
		if cfg.LFile == "" {
			logger.Fatal("Missing -Lf (media file)")
			os.Exit(1)
		}

		if mode != "stream" {
			if err := utils.ValidateFile(cfg.LFile); err != nil {
				logger.Fatal("Invalid file: %v", err)
				os.Exit(1)
			}
		}

		cfg.LIP = utils.LocalIP(cfg.LIP)
		if mode != "scan" && !cfg.ProbeOnly {
			if mode != "stream" {
				internal.ServeDirGo(cfg, stop)
			}
		}
		time.Sleep(500 * time.Millisecond)
		serverRunning = true
	}

	internal.RunScript(&cfg, stop)

	if !serverRunning {
		return
	}

	logger.Info("renderctl running — press Ctrl+C to exit")
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	<-sig
	close(stop)
}
