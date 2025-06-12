package main

import (
	"github.com/Soliard/go-tpl-metrics/internal/agent"
	"github.com/Soliard/go-tpl-metrics/internal/misc"
)

func main() {
	agent := agent.NewAgent(misc.GetAgentURL())
	agent.Run()
}
