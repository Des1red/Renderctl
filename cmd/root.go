package cmd

import (
	"os"
	"os/signal"
	"time"

	"renderctl/internal"
	"renderctl/internal/cache"
	"renderctl/internal/models"
	"renderctl/internal/servers"
	"renderctl/internal/stream"
	"renderctl/internal/ui"
	"renderctl/internal/utils"
	"renderctl/logger"
	"renderctl/requirements"
)

var cfg = models.DefaultConfig
var noCache bool

func Execute() {
	parseFlags()
	// ---- INSTALLER (early exit) ----
	if requirements.Install {
		if err := requirements.RunInstaller(); err != nil {
			logger.Error("%v", err)
		}
		os.Exit(0)
	}

	if bad, msg := badFlagUse(); bad {
		logger.Error(msg)
	}
	// Set verbose
	logger.SetVerbose(cfg.Verbose)
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
			logger.Error("Missing -Lf (media file)")
		}

		if mode != "stream" {
			if err := utils.ValidateFile(cfg.LFile); err != nil {
				logger.Error("Invalid file: %v", err)
			}
		}

		cfg.LIP = utils.LocalIP(cfg.LIP)
		if mode != "scan" && !cfg.ProbeOnly {
			if mode != "stream" {
				servers.InitDefaultServer(cfg, stop)
			} else if mode == "stream" {
				stream.InitStreamServer(&cfg, stop)
			}
		}
		time.Sleep(500 * time.Millisecond)
		serverRunning = true
	}

	internal.RunScript(&cfg)

	if !serverRunning {
		return
	}

	logger.Status("renderctl running — press Ctrl+C to exit")
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	<-sig
	close(stop)
}
