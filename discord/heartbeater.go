package discord

import (
	"fmt"
	"garrison-stauffer.com/discord-bot/discord/gateway"
	"log"
	"time"
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
		// restart it
		log.Println("Stopping current heartbeater")
		err := h.Stop()
		if err != nil {
			fmt.Println("Couldn't stop the heartbeater, going to skip starting a new one and hope for the best")
			return
		}
	}
	h.isRunning = true

	ticker := time.NewTicker(interval)

	for {
		select {
		case <-ticker.C:
			fmt.Println(h.client.gatewayConfig.sequence)
			msg := gateway.NewHeartbeat(h.client.gatewayConfig.sequence)
			log.Printf("sending heartbeat")
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
