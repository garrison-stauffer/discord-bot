package discord

import (
	"fmt"
	"log/slog"
	"time"

	"garrison-stauffer.com/discord-bot/discord/gateway"
)

type Heartbeater interface {
	Start(interval time.Duration)
	Stop() error
}

type heartbeaterImpl struct {
	isRunning            bool
	stopBeating          chan struct{}
	notifyStoppedBeating chan struct{}
	client               *clientImpl
}

func NewHeartbeater(c *clientImpl) Heartbeater {
	return &heartbeaterImpl{
		stopBeating:          make(chan struct{}),
		notifyStoppedBeating: make(chan struct{}),
		client:               c,
	}
}

func (h *heartbeaterImpl) Start(interval time.Duration) {
	if h.isRunning {
		slog.Info("stopping current heartbeater to restart")
		err := h.Stop()
		if err != nil {
			slog.Error("couldn't stop heartbeater, skipping restart", "error", err)
			return
		}
	}
	h.isRunning = true

	ticker := time.NewTicker(interval)

	for {
		select {
		case <-ticker.C:
			msg := gateway.NewHeartbeat(h.client.gatewayConfig.sequence)
			slog.Debug("sending heartbeat", "sequence", h.client.gatewayConfig.sequence)
			_ = h.client.Send(*msg)

		case <-h.stopBeating:
			ticker.Stop()
			h.notifyStoppedBeating <- struct{}{}
			return
		}
	}
}

func (h *heartbeaterImpl) Stop() error {
	h.stopBeating <- struct{}{}

	select {
	case <-h.notifyStoppedBeating:
		return nil
	case <-time.After(5 * time.Second):
		return fmt.Errorf("couldn't stop beating")
	}
}
