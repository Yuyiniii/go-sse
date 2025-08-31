package heartbeat

import (
	"sse/internal/ports"
	"time"
)

// Heartbeat 控制心跳消息的发送
type Heartbeat struct {
	ticker *time.Ticker
	hub    ports.Hub
}

// NewHeartbeat 创建一个新的 Heartbeat 实例
func NewHeartbeat(interval int, hub ports.Hub) *Heartbeat {
	return &Heartbeat{
		ticker: time.NewTicker(time.Duration(interval) * time.Second),
		hub:    hub,
	}
}

func (h *Heartbeat) Start() {
	go func() {
		for {
			select {
			case <-h.ticker.C:
				bytes := []byte("event: ping")
				h.hub.HeaderBeat(bytes)

			}
		}
	}()
}
