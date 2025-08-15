package acct

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/marcboeker/go-duckdb/v2"
	"github.com/rs/zerolog/log"
)

func Query(ctx context.Context, dbPath string, queryInputs []string) ([]map[string]interface{}, error) {
	log.Info().Str("db", dbPath).Msg("querying account")
	db, err := sql.Open("duckdb", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open duckdb: %w", err)
	}
	defer db.Close()

	var allResults []map[string]interface{}

	if len(queryInputs) == 0 {
		return nil, fmt.Errorf("no query names specified")
	}

	for _, queryInput := range queryInputs {
		var results []map[string]interface{}
		var err error
		if strings.Contains(queryInput, " ") {
			results, err = runRawSQL(ctx, db, "adhoc", queryInput)
		} else {
			results, err = runSQLFile(ctx, db, "query", queryInput)
		}
		if err != nil {
			return nil, err
		}
		allResults = append(allResults, results...)
	}

	return allResults, nil
}
