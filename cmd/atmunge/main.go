package main

import (
	"log"
	_ "net/http/pprof"
	"os"

	"github.com/blebbit/atmunge/cmd/at-mirror/cmd"
	"github.com/blebbit/atmunge/pkg/config"
)

func main() {
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("failed to load config: %s", err)
		os.Exit(1)
	}

	cmd.Execute()
}
