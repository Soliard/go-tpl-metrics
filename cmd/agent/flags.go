package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/Soliard/go-tpl-metrics/internal/agent"
	"github.com/caarlos0/env/v6"
)

func ParseFlags() agent.Config {
	config := agent.Config{}

	flag.StringVar(&config.ServerHost, "a", "localhost:8080", "server addres")
	flag.DurationVar(&config.PollInterval, "p", 2*time.Second, "metrics poll interval is seconds")
	flag.DurationVar(&config.ReportInterval, "r", 10*time.Second, "metrics send interval in seconds")
	flag.Parse()

	fmt.Println(`agent parsed flags config: `, config)

	err := env.Parse(&config)
	if err != nil {
		fmt.Println(`cannot parse config from env for agent`)
	}

	fmt.Println(`agent after env config: `, config)

	return config
}
