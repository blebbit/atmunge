package cmd

import (
	"context"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/blebbit/at-mirror/pkg/config"
	"github.com/blebbit/at-mirror/pkg/runtime"
	"github.com/blebbit/at-mirror/pkg/server"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(runCmd)
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the at-mirror in sync & server mode",
	Long:  "Run the at-mirror in sync & server mode",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer stop()

		ctx, err := config.SetupLogging(ctx)
		if err != nil {
			return err
		}
		log := zerolog.Ctx(ctx).With().
			Str("module", "server").
			Str("method", "run").
			Logger()

		// create our runtime
		r, err := runtime.NewRuntime(ctx)
		if err != nil {
			log.Error().Msgf("failed to create runtime: %s", err)
			return err
		}

		// (maybe) start mirror
		if r.Cfg.RunPlcMirror {
			go func() {
				r.StartPLCMirror()
			}()
		}

		s := server.NewServer(r)
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

		return nil
	},
}

func runMain() error {

	// // create our runtime
	// r, err := runtime.NewRuntime(ctx, db)
	// if err != nil {
	// 	return fmt.Errorf("failed to create runtime: %w", err)
	// }

	// var wg sync.WaitGroup

	// if config.RunRepoMirror {
	// 	wg.Add(1)
	// 	go func() {
	// 		defer wg.Done()
	// 		r.StartRepoMirror()
	// 	}()
	// }

	return nil
}
