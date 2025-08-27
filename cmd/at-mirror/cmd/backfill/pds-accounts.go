package backfill

import (
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"syscall"

	"github.com/blebbit/atmunge/pkg/config"
	"github.com/blebbit/atmunge/pkg/runtime"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

var genListStartPDS string // PDS to start from, if empty, starts from the first PDS in the list

func init() {
	BackfillCmd.AddCommand(backfillPdsAccountsCmd)
}

var backfillPdsAccountsCmd = &cobra.Command{
	Use:   "pds-accounts",
	Short: "Backfill the list of repos per PDS",
	Long:  "Backfill the list of repos per PDS",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer stop()

		ctx, err := config.SetupLogging(ctx)
		if err != nil {
			return err
		}
		log := zerolog.Ctx(ctx).With().
			Str("module", "backfill").
			Str("method", "pds-accounts").
			Logger()
		log.Info().Msgf("Starting up...")

		// create our runtime
		r, err := runtime.NewRuntime(ctx)
		if err != nil {
			log.Error().Msgf("failed to create runtime: %s", err)
			return err
		}

		// load repo list from json (TODO, should put this in a table for consistency, have some command to fetch and sync)
		j, err := os.ReadFile("./data/atproto-scraping-state.json")
		if err != nil {
			log.Error().Msgf("failed to read json file: %s", err)
			return err
		}

		d := map[string]map[string]interface{}{}
		if err := json.Unmarshal(j, &d); err != nil {
			log.Error().Msgf("failed to unmarshal json: %s", err)
			return err
		}
		p := d["pdses"]

		pdses := make([]string, 0, len(p))
		for url, val := range p {
			pp := val.(map[string]interface{})
			if _, ok := pp["errorAt"]; ok {
				continue
			}
			pdses = append(pdses, url)
		}

		log.Info().Msgf("Found %d PDSes", len(pdses))

		err = r.BackfillPdsAccounts(pdses, genListStartPDS)
		if err != nil {
			log.Error().Msgf("failed to backfill PLC logs: %s", err)
			return err
		}

		return nil
	},
}
