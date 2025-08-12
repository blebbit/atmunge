package acct

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/marcboeker/go-duckdb/v2"
	"github.com/rs/zerolog/log"
)

func Query(ctx context.Context, dbPath string, queryNames []string) ([]map[string]interface{}, error) {
	log.Info().Str("db", dbPath).Msg("querying account")
	db, err := sql.Open("duckdb", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open duckdb: %w", err)
	}
	defer db.Close()

	var allResults []map[string]interface{}

	if len(queryNames) == 0 {
		return nil, fmt.Errorf("no query names specified")
	}

	for _, queryName := range queryNames {
		results, err := runSQLFile(ctx, db, "query", queryName)
		if err != nil {
			return nil, err
		}
		allResults = append(allResults, results...)
	}

	return allResults, nil
}
