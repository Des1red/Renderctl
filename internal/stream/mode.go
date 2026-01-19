package stream

import (
	"time"

	"renderctl/internal/avtransport"
	"renderctl/internal/models"
	"renderctl/internal/servers"
	"renderctl/internal/utils"
	"renderctl/logger"
)

var runtimePlan *StreamPlan

func SetRuntimePlan(p *StreamPlan) {
	runtimePlan = p
}

func StartStreamServer(
	cfg *models.Config,
	plan *StreamPlan,
	stop <-chan struct{},
) {
	servers.ServeStreamGo(
		cfg,
		stop,
		plan.StreamPath,
		plan.Mime,
		plan.Container,
		plan.Source,
	)

	for !cfg.ServerUp {
		time.Sleep(100 * time.Millisecond)
	}
}

func InitStreamServer(cfg *models.Config, stop <-chan struct{}) {
	plan, err := ResolveStreamPlan(cfg)
	if err != nil {
		logger.Fatal("Stream setup failed: %v", err)
		return
	}

	SetRuntimePlan(plan)
	StartStreamServer(cfg, plan, stop)
}

func StartStreamPlay(cfg *models.Config) {
	if runtimePlan == nil {
		logger.Fatal("StreamPlan missing (internal state error)")
		return
	}

	controlURL := utils.ControlURL(cfg)
	if cfg.CachedControlURL != "" {
		controlURL = cfg.CachedControlURL
	}

	target := avtransport.Target{
		ControlURL: controlURL,
		MediaURL:   BuildStreamURL(cfg, runtimePlan.StreamPath),
	}

	avtransport.Run(target, "")
}
