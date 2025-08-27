package main

import (
	"github.com/obsidian-agent-cli/internal/client"
	"github.com/obsidian-agent-cli/internal/property"
)

func main(){
	cfg := property.GetDefaultConfig()
	client.Run(cfg)
}