package main

import (
	"fmt"

	"github.com/Soliard/go-tpl-metrics/internal/agent"
)

func main() {
	config := ParseFlags()
	agent := agent.NewAgent(config)
	fmt.Println("agent works with service on ", config.ServerHost)
	agent.Run()
}
