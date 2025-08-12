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

func (i *Indexer) Index(ctx context.Context, dbPath string, indexName string) error {
	log.Info().Str("db", dbPath).Msg("indexing account")
	db, err := sql.Open("duckdb", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open duckdb: %w", err)
	}
	defer db.Close()

	if indexName != "" {
		return i.runSQLFile(ctx, db, indexName)
	}

	return i.runAllSQLFiles(ctx, db)
}

func (i *Indexer) runSQLFile(ctx context.Context, db *sql.DB, fileName string) error {
	sql, err := sqlFiles.ReadFile(filepath.Join("sql", fileName))
	if err != nil {
		return fmt.Errorf("failed to read sql file %s: %w", fileName, err)
	}

	log.Info().Str("file", fileName).Msg("executing sql file")
	_, err = db.ExecContext(ctx, string(sql))
	if err != nil {
		return fmt.Errorf("failed to execute sql file %s: %w", fileName, err)
	}

	return nil
}

func (i *Indexer) runAllSQLFiles(ctx context.Context, db *sql.DB) error {
	files, err := fs.Glob(sqlFiles, "sql/*.sql")
	if err != nil {
		return fmt.Errorf("failed to glob sql files: %w", err)
	}
	sort.Strings(files)

	for _, file := range files {
		baseName := filepath.Base(file)
		if err := i.runSQLFile(ctx, db, baseName); err != nil {
			return err
		}
	}

	return nil
}
