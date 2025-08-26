package orchestrator

import (
	"context"

	"github.com/obsidian-agent/internal/transport"
)

type Orchestrator interface {
	Run(ctx context.Context, msg transport.Msg, sender transport.Sender) error
	Cancel(id string)
}