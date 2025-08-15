package acct

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"

	appsql "github.com/blebbit/at-mirror/pkg/sql"
	"github.com/rs/zerolog/log"
)

func runSQLFile(ctx context.Context, db *sql.DB, queryType, fileName string) ([]map[string]interface{}, error) {
	if !strings.HasSuffix(fileName, ".sql") {
		fileName += ".sql"
	}
	sqlBytes, err := appsql.SQLFiles.ReadFile(filepath.Join("acct", queryType, fileName))
	if err != nil {
		return nil, fmt.Errorf("failed to read sql file %s: %w", fileName, err)
	}

	log.Info().Str("file", fileName).Str("type", queryType).Msg("executing sql file")
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

func runAllSQLFiles(ctx context.Context, db *sql.DB, queryType string) ([]map[string]interface{}, error) {
	files, err := fs.Glob(appsql.SQLFiles, filepath.Join("acct", queryType, "*.sql"))
	if err != nil {
		return nil, fmt.Errorf("failed to glob sql files: %w", err)
	}
	sort.Strings(files)

	var allResults []map[string]interface{}
	for _, file := range files {
		baseName := filepath.Base(file)
		results, err := runSQLFile(ctx, db, queryType, baseName)
		if err != nil {
			return nil, err
		}
		allResults = append(allResults, results...)
	}

	return allResults, nil
}

func runRawSQL(ctx context.Context, db *sql.DB, queryType, rawSQL string) ([]map[string]interface{}, error) {
	log.Info().Str("sql", rawSQL).Str("type", queryType).Msg("executing raw sql")
	rows, err := db.QueryContext(ctx, rawSQL)
	if err != nil {
		// If it's a DDL statement, it might not return rows.
		if _, execErr := db.ExecContext(ctx, rawSQL); execErr == nil {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to execute raw sql: %w", err)
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
