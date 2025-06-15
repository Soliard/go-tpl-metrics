package main

import (
	"flag"
	"fmt"

	"github.com/Soliard/go-tpl-metrics/internal/server"
	"github.com/caarlos0/env/v6"
)

func ParseFlags() server.Config {
	config := server.Config{}

	flag.StringVar(&config.ServerHost, "a", "localhost:8080", "server addres")
	flag.Parse()

	fmt.Println(`server parsed flags config: `, config)

	err := env.Parse(&config)
	if err != nil {
		fmt.Println(`cannot parse config from env for server`)
	}

	fmt.Println(`server after env config: `, config)

	return config
}
