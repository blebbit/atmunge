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

func CarToDuckDB(carPath string, dbPath string) error {
	f, err := os.Open(carPath)
	if err != nil {
		return fmt.Errorf("failed to open car file: %w", err)
	}
	defer f.Close()

	r, err := ReadRepoFromCar(f)
	if err != nil {
		return fmt.Errorf("failed to read repo from car: %w", err)
	}

	return RepoToDuckDB(r, dbPath)
}

func RepoToDuckDB(r *indigoRepo.Repo, dbPath string) error {
	db, err := InitDuckDB(dbPath)
	if err != nil {
		return fmt.Errorf("failed to init duckdb: %w", err)
	}
	defer db.Close()

	return SaveRecordsToDuckDB(context.Background(), r, db)
}

func BlockstoreToDuckDB(ctx context.Context, bsMap map[cid.Cid][]byte, root cid.Cid, dbPath string) error {
	r, err := BlockstoreToRepo(ctx, bsMap, root)
	if err != nil {
		return fmt.Errorf("failed to convert blockstore to repo: %w", err)
	}
	return RepoToDuckDB(r, dbPath)
}

func SaveNewRecordsToDuckDB(ctx context.Context, bsMap map[cid.Cid][]byte, newBlocks map[cid.Cid][]byte, root cid.Cid, dbPath string) error {
	r, err := BlockstoreToRepo(ctx, bsMap, root)
	if err != nil {
		return fmt.Errorf("failed to convert blockstore to repo: %w", err)
	}

	db, err := InitDuckDB(dbPath)
	if err != nil {
		return fmt.Errorf("failed to init duckdb: %w", err)
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO records (created_at, indexed_at, updated_at, did, nsid, rkey, cid, record)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?);
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	// get the DID from the MST here

	err = r.MST.Walk(func(k []byte, v cid.Cid) error {
		if _, isNew := newBlocks[v]; !isNew {
			return nil
		}

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
		if ca, ok := rec["updatedAt"].(string); ok {
			if t, err := time.Parse(time.RFC3339, ca); err == nil {
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

		_, err = stmt.Exec(createdAt, indexedAt, updatedAt, col.String(), rkey.String(), v.String(), string(recJSON))
		return err
	})

	if err != nil {
		return err
	}

	return tx.Commit()
}

func SaveRecordsToDuckDB(ctx context.Context, r *indigoRepo.Repo, db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO records (created_at, indexed_at, nsid, rkey, cid, record)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT (nsid, rkey, cid) DO UPDATE SET
		created_at = excluded.created_at,
		indexed_at = excluded.indexed_at,
		record = excluded.record;
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

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

		var createdAt, indexedAt time.Time
		if ca, ok := rec["createdAt"].(string); ok {
			if t, err := time.Parse(time.RFC3339, ca); err == nil {
				createdAt = t
			}
		}
		if ia, ok := rec["indexedAt"].(string); ok {
			if t, err := time.Parse(time.RFC3339, ia); err == nil {
				indexedAt = t
			}
		} else {
			indexedAt = time.Now()
		}

		_, err = stmt.Exec(createdAt, indexedAt, col.String(), rkey.String(), v.String(), string(recJSON))
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
