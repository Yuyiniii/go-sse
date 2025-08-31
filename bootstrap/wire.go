package bootstrap

import (
	"sse/internal/adapters/hub"
	"sse/internal/ports"
)

type Container struct {
	ShardedHub ports.Hub
}

func NewContainer() *Container {
	return &Container{
		ShardedHub: hub.NewShardedHub(),
	}
}
