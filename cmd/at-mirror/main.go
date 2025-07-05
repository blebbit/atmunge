package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/rs/zerolog"

	atdb "github.com/blebbit/at-mirror/pkg/db"
	atrt "github.com/blebbit/at-mirror/pkg/runtime"
	atsrv "github.com/blebbit/at-mirror/pkg/server"
)

func main() {
	if err := loadConfig(); err != nil {
		log.Fatalf("failed to load config: %s", err)
	}

	return

	if err := runMain(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runMain() error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	ctx = setupLogging(ctx)
	log := zerolog.Ctx(ctx)
	log.Debug().Msgf("Starting up...")

	// db setup
	db, err := atdb.GetClient(config.DBUrl, ctx)
	if err != nil {
		return err
	}
	log.Debug().Msgf("DB connection established")

	// db migrations (if needed)
	err = atdb.MigrateModels(db)
	if err != nil {
		return err
	}
	log.Debug().Msgf("DB schema updated")

	// create our runtime
	r, err := atrt.NewRuntime(ctx, db)
	if err != nil {
		return fmt.Errorf("failed to create runtime: %w", err)
	}

	var wg sync.WaitGroup

	// (maybe) start mirror
	if config.RunPlcMirror {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r.StartPLCMirror()
		}()
	}

	if config.RunRepoMirror {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r.StartRepoMirror()
		}()
	}

	if config.RunServer {
		s := atsrv.NewServer(r)
		// start server
		log.Info().Msgf("Starting HTTP listener on %q...", ":1323")

		go func() {
			if err := s.Echo().Start(":1323"); err != nil && err != http.ErrServerClosed {
				s.Echo().Logger.Fatal("shutting down the server")
			}
		}()

		// Wait for interrupt signal to gracefully shut down the server with a timeout of 10 seconds.
		<-ctx.Done()
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := s.Echo().Shutdown(ctx); err != nil {
			s.Echo().Logger.Fatal(err)
		}
	} else {
		// if not running the server, we are waiting for the mirrors to finish
		log.Info().Msgf("Not running the server, waiting for mirrors to finish...")
		wg.Wait()
		log.Info().Msgf("All mirrors finished, exiting.")
	}

	return nil
}
