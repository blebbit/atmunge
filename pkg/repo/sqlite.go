package repo

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/bluesky-social/indigo/atproto/data"
	"github.com/bluesky-social/indigo/atproto/syntax"
	"github.com/ipfs/go-cid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	indigoRepo "github.com/bluesky-social/indigo/atproto/repo"
)

type Record struct {
	CreatedAt time.Time `gorm:"created_at"`
	UpdatedAt time.Time `gorm:"updated_at"`
	IndexedAt time.Time `gorm:"indexed_at"`
	NSID      string    `gorm:"column:nsid;primaryKey"`
	RKey      string    `gorm:"column:rkey;primaryKey"`
	CID       string    `gorm:"column:cid;primaryKey"`
	Record    string    `gorm:"record"`
}

func (Record) TableName() string {
	return "records"
}

func InitSQLite(dbPath string) (*gorm.DB, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create sqlite directory: %w", err)
	}

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite db: %w", err)
	}

	if err := db.AutoMigrate(&Record{}); err != nil {
		return nil, fmt.Errorf("failed to migrate sqlite db: %w", err)
	}

	return db, nil
}

func CarToSQLite(carPath string, dbPath string) error {
	f, err := os.Open(carPath)
	if err != nil {
		return fmt.Errorf("failed to open car file: %w", err)
	}
	defer f.Close()

	r, err := ReadRepoFromCar(f)
	if err != nil {
		return fmt.Errorf("failed to read repo from car: %w", err)
	}

	return RepoToSQLite(r, dbPath)
}

func RepoToSQLite(r *indigoRepo.Repo, dbPath string) error {
	db, err := InitSQLite(dbPath)
	if err != nil {
		return fmt.Errorf("failed to init sqlite: %w", err)
	}

	return SaveRecordsToSQLite(context.Background(), r, db)
}

func BlockstoreToSQLite(ctx context.Context, bsMap map[cid.Cid][]byte, root cid.Cid, dbPath string) error {
	r, err := BlockstoreToRepo(ctx, bsMap, root)
	if err != nil {
		return fmt.Errorf("failed to convert blockstore to repo: %w", err)
	}
	return RepoToSQLite(r, dbPath)
}

func SaveRecordsToSQLite(ctx context.Context, r *indigoRepo.Repo, db *gorm.DB) error {
	return r.MST.Walk(func(k []byte, v cid.Cid) error {
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

		dbRec := Record{
			IndexedAt: time.Now(),
			NSID:      col.String(),
			RKey:      rkey.String(),
			CID:       v.String(),
			Record:    string(recJSON),
		}

		// Attempt to extract createdAt from the record
		if ca, ok := rec["createdAt"].(string); ok {
			if t, err := time.Parse(time.RFC3339, ca); err == nil {
				dbRec.CreatedAt = t
			}
		}

		// Attempt to extract indexedAt from the record
		if ca, ok := rec["indexedAt"].(string); ok {
			if t, err := time.Parse(time.RFC3339, ca); err == nil {
				dbRec.IndexedAt = t
			}
		}

		return db.Save(&dbRec).Error
	})
}
