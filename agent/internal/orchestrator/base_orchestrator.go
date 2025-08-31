package orchestrator

import (
	"context"

	"github.com/obsidian-agent/biz/transport"
)

type Orchestrator interface {
	Run(ctx context.Context, msg transport.MsgRequest, sender transport.Sender) error
	Cancel(id string)
}