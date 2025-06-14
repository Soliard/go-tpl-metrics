package main

import (
	"flag"

	"github.com/Soliard/go-tpl-metrics/internal/server"
)

func ParseFlags() server.Config {
	config := server.Config{}

	flag.StringVar(&config.ServerHost, "a", "localhost:8080", "server addres")
	flag.Parse()

	return config
}
