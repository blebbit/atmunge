package acct

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"

	_ "github.com/marcboeker/go-duckdb/v2"
	"github.com/rs/zerolog/log"
)

//go:embed sql/*.sql
var sqlFiles embed.FS

type Indexer struct {
}

func NewIndexer() *Indexer {
	return &Indexer{}
}

func (i *Indexer) Index(ctx context.Context, dbPath string, indexName string) ([]map[string]interface{}, error) {
	log.Info().Str("db", dbPath).Msg("indexing account")
	db, err := sql.Open("duckdb", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open duckdb: %w", err)
	}
	defer db.Close()

	if indexName != "" {
		return i.runSQLFile(ctx, db, indexName)
	}

	return i.runAllSQLFiles(ctx, db)
}

func (i *Indexer) runSQLFile(ctx context.Context, db *sql.DB, fileName string) ([]map[string]interface{}, error) {
	sqlBytes, err := sqlFiles.ReadFile(filepath.Join("sql", fileName))
	if err != nil {
		return nil, fmt.Errorf("failed to read sql file %s: %w", fileName, err)
	}

	log.Info().Str("file", fileName).Msg("executing sql file")
	rows, err := db.QueryContext(ctx, string(sqlBytes))
	if err != nil {
		// If it's a DDL statement, it might not return rows.
		if _, execErr := db.ExecContext(ctx, string(sqlBytes)); execErr == nil {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to execute sql file %s: %w", fileName, err)
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	var results []map[string]interface{}
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		rowData := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			switch v := val.(type) {
			case []byte:
				rowData[col] = string(v)
			default:
				rowData[col] = v
			}
		}
		results = append(results, rowData)
	}

	return results, nil
}

func (i *Indexer) runAllSQLFiles(ctx context.Context, db *sql.DB) ([]map[string]interface{}, error) {
	files, err := fs.Glob(sqlFiles, "sql/*.sql")
	if err != nil {
		return nil, fmt.Errorf("failed to glob sql files: %w", err)
	}
	sort.Strings(files)

	var allResults []map[string]interface{}
	for _, file := range files {
		baseName := filepath.Base(file)
		results, err := i.runSQLFile(ctx, db, baseName)
		if err != nil {
			return nil, err
		}
		allResults = append(allResults, results...)
	}

	return allResults, nil
}
