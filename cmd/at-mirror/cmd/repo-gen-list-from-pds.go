package cmd

import (
	"context"
	"encoding/json"
	"os"

	"github.com/blebbit/at-mirror/pkg/config"
	"github.com/blebbit/at-mirror/pkg/db"
	"github.com/blebbit/at-mirror/pkg/runtime"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

var genListStartPDS string // PDS to start from, if empty, starts from the first PDS in the list

func init() {
	repoCmd.AddCommand(repoGenListFromPdsCmd)
	repoGenListFromPdsCmd.Flags().StringVar(&genListStartPDS, "start", "", "Continue from this PDS")
}

var repoGenListFromPdsCmd = &cobra.Command{
	Use:   "gen-list-from-pds",
	Short: "Generate a list of repos from PDS list",
	Long:  "Generate a list of repos from PDS list",
	RunE: func(cmd *cobra.Command, args []string) error {
		// ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		// defer stop()
		ctx := context.Background()

		ctx = config.SetupLogging(ctx)
		log := zerolog.Ctx(ctx).With().Str("module", "repo").Logger()
		log.Info().Msgf("Starting up...")

		cfg := config.GetConfig()

		// load repo list from json
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

		// db setup
		DB, err := db.GetClient(cfg.DBUrl, ctx)
		if err != nil {
			return err
		}
		log.Info().Msgf("DB connection established")

		// create our runtime
		r, err := runtime.NewRuntime(ctx, DB)
		if err != nil {
			log.Error().Msgf("failed to create runtime: %s", err)
			return err
		}

		err = r.RepoListFromPDS(pdses, genListStartPDS)
		if err != nil {
			log.Error().Msgf("failed to backfill PLC logs: %s", err)
			return err
		}

		return nil
	},
}
