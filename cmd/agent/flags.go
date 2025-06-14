package main

import (
	"flag"

	"github.com/Soliard/go-tpl-metrics/internal/agent"
)

func ParseFlags() agent.Config {
	config := agent.Config{}

	flag.StringVar(&config.ServerHost, "a", "localhost:8080", "server addres")
	flag.IntVar(&config.PollIntervalSeconds, "p", 2, "metrics poll interval is seconds")
	flag.IntVar(&config.ReportIntervalSeconds, "r", 10, "metrics send interval in seconds")
	flag.Parse()

	return config
}
