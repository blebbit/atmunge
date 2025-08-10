package repo

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/bluesky-social/indigo/atproto/data"
	indigoRepo "github.com/bluesky-social/indigo/atproto/repo"
	"github.com/bluesky-social/indigo/atproto/syntax"
	"github.com/ipfs/go-cid"
	_ "github.com/marcboeker/go-duckdb"
)

func InitDuckDB(dbPath string) (*sql.DB, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create duckdb directory: %w", err)
	}

	db, err := sql.Open("duckdb", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open duckdb: %w", err)
	}

	createTableSQL := `
	CREATE TABLE IF NOT EXISTS records (
		created_at TIMESTAMP,
		updated_at TIMESTAMP,
		indexed_at TIMESTAMP,
		nsid VARCHAR,
		rkey VARCHAR,
		cid VARCHAR,
		record JSON,
		PRIMARY KEY (nsid, rkey, cid)
	);`
	if _, err := db.Exec(createTableSQL); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create records table: %w", err)
	}

	return db, nil
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
