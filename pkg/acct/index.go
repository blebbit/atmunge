package acct

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/marcboeker/go-duckdb/v2"
	"github.com/rs/zerolog/log"
)

func Index(ctx context.Context, dbPath string, indexNames []string) error {
	log.Info().Str("db", dbPath).Msg("indexing account")
	db, err := sql.Open("duckdb", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open duckdb: %w", err)
	}
	defer db.Close()

	if len(indexNames) == 0 {
		// Default behavior: run all index queries
		_, err := runAllSQLFiles(ctx, db, "index")
		if err != nil {
			return err
		}
	} else {
		for _, indexName := range indexNames {
			_, err := runSQLFile(ctx, db, "index", indexName)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
