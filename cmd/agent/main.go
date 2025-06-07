package main

import "github.com/Soliard/go-tpl-metrics/internal/agent"

func main() {
	agent := agent.NewAgent(`http://localhost:8080`)
	agent.Run()
}
