package repo

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/blebbit/at-mirror/pkg/db"
	"github.com/bluesky-social/indigo/atproto/data"
	indigoRepo "github.com/bluesky-social/indigo/atproto/repo"
	"github.com/bluesky-social/indigo/atproto/syntax"
	"github.com/ipfs/go-cid"
	_ "github.com/marcboeker/go-duckdb/v2"
	"github.com/nrednav/cuid2"
)

func InitDuckDB(dbPath string) (*sql.DB, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create duckdb directory: %w", err)
	}

	dbConn, err := sql.Open("duckdb", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open duckdb: %w", err)
	}

	// get the path to the migrations directory
	// HACK: this is not a good way to do this
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current working directory: %w", err)
	}
	migrationsPath := filepath.Join(cwd, "pkg", "sql", "acct", "migrations")

	if err := db.RunDuckDBMigrations(dbConn, migrationsPath); err != nil {
		dbConn.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return dbConn, nil
}

func CarToDuckDB(ctx context.Context, carPath string, dbPath string) error {
	// Initialize the database first
	db, err := InitDuckDB(dbPath)
	if err != nil {
		return fmt.Errorf("failed to init duckdb: %w", err)
	}
	defer db.Close()

	// Open and read the CAR file
	f, err := os.Open(carPath)
	if err != nil {
		return fmt.Errorf("failed to open car file: %w", err)
	}
	defer f.Close()

	r, err := ReadRepoFromCar(f)
	if err != nil {
		return fmt.Errorf("failed to read repo from car: %w", err)
	}

	// Save the records to the database
	return SaveRecordsToDuckDB(ctx, r, db)
}

func RepoToDuckDB(r *indigoRepo.Repo, dbPath string) error {
	db, err := InitDuckDB(dbPath)
	if err != nil {
		return fmt.Errorf("failed to init duckdb: %w", err)
	}
	defer db.Close()

	return SaveRecordsToDuckDB(context.Background(), r, db)
}

func SaveRecordsToDuckDB(ctx context.Context, r *indigoRepo.Repo, db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO records (cuid, created_at, indexed_at, updated_at, did, nsid, rkey, cid, record)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(did, nsid, rkey, cid) DO NOTHING;
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	repoDid := r.DID.String()

	err = r.MST.Walk(func(k []byte, v cid.Cid) error {
		col, rkey, err := syntax.ParseRepoPath(string(k))
		if err != nil {
			return err
		}

		recBytes, _, err := r.GetRecordBytes(ctx, col, rkey)
		if err != nil {
			return err
		}

		rec, err := data.UnmarshalCBOR(recBytes)
		if err != nil {
			return err
		}

		recJSON, err := json.Marshal(rec)
		if err != nil {
			return err
		}

		var createdAt, updatedAt, indexedAt time.Time
		if ca, ok := rec["createdAt"].(string); ok {
			if t, err := time.Parse(time.RFC3339, ca); err == nil {
				createdAt = t
			}
		}
		if ua, ok := rec["updatedAt"].(string); ok {
			if t, err := time.Parse(time.RFC3339, ua); err == nil {
				updatedAt = t
			}
		}
		if ia, ok := rec["indexedAt"].(string); ok {
			if t, err := time.Parse(time.RFC3339, ia); err == nil {
				indexedAt = t
			}
		} else {
			indexedAt = time.Now()
		}

		cuid := cuid2.Generate()

		_, err = stmt.Exec(cuid, createdAt, indexedAt, updatedAt, repoDid, col.String(), rkey.String(), v.String(), string(recJSON))
		return err
	})

	if err != nil {
		return err
	}

	return tx.Commit()
}

func GetRecord(dbPath, nsid, rkey string) (json.RawMessage, error) {
	db, err := InitDuckDB(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to init duckdb: %w", err)
	}
	defer db.Close()

	var recordJSON string
	query := "SELECT record::TEXT FROM records WHERE nsid = ? AND rkey = ?"
	err = db.QueryRow(query, nsid, rkey).Scan(&recordJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no record found with nsid %s and rkey %s", nsid, rkey)
		}
		return nil, fmt.Errorf("failed to query record: %w", err)
	}

	return json.RawMessage(recordJSON), nil
}

func ClearDuckDBTables(db *sql.DB, tables []string) error {
	if len(tables) == 0 {
		// If no tables are specified, get all tables from the database
		rows, err := db.Query("SHOW TABLES")
		if err != nil {
			return fmt.Errorf("failed to query tables: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var name string
			if err := rows.Scan(&name); err != nil {
				return fmt.Errorf("failed to scan table name: %w", err)
			}
			tables = append(tables, name)
		}
	}

	for _, table := range tables {
		if _, err := db.Exec("DELETE FROM " + table); err != nil {
			return fmt.Errorf("clearing table %q: %w", table, err)
		}
	}

	return nil
}
