package main

import (
	"log"
	_ "net/http/pprof"
	"os"

	"github.com/blebbit/at-mirror/cmd/at-mirror/cmd"
	"github.com/blebbit/at-mirror/pkg/config"
)

func main() {
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("failed to load config: %s", err)
		os.Exit(1)
	}

	cmd.Execute()
}
