package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/blebbit/at-mirror/pkg/util/gormzerolog"
)

var dbSingleton *gorm.DB

func GetClient(dbUrl string, ctx context.Context) (*gorm.DB, error) {
	if dbSingleton != nil {
		return dbSingleton, nil
	}

	dbCfg, err := pgxpool.ParseConfig(dbUrl)
	if err != nil {
		return nil, fmt.Errorf("parsing DB URL: %w", err)
	}
	dbCfg.MaxConns = 4096
	dbCfg.MinConns = 10
	dbCfg.MaxConnLifetime = 6 * time.Hour
	conn, err := pgxpool.NewWithConfig(ctx, dbCfg)
	if err != nil {
		return nil, fmt.Errorf("connecting to postgres: %w", err)
	}

	sqldb := stdlib.OpenDBFromPool(conn)

	db, err := gorm.Open(postgres.New(postgres.Config{
		Conn: sqldb,
	}), &gorm.Config{
		Logger: gormzerolog.New(&logger.Config{
			SlowThreshold:             10 * time.Second,
			IgnoreRecordNotFoundError: true,
		}, nil),
	})

	if err != nil {
		return nil, fmt.Errorf("connecting to the database: %w", err)
	}

	dbSingleton = db

	return db, nil
}

func MigrateModels(db *gorm.DB) error {
	if err := db.AutoMigrate(&PLCLogEntry{}); err != nil {
		return fmt.Errorf("auto-migrating DB schema: %w", err)
	}
	if err := db.AutoMigrate(&AccountInfo{}); err != nil {
		return fmt.Errorf("auto-migrating DB schema: %w", err)
	}
	if err := db.AutoMigrate(&PdsRepo{}); err != nil {
		return fmt.Errorf("auto-migrating DB schema: %w", err)
	}
	if err := db.AutoMigrate(&AccountRepo{}); err != nil {
		return fmt.Errorf("auto-migrating DB schema: %w", err)
	}

	return nil
}

func ClearTables(db *gorm.DB, tables []string) error {
	if len(tables) == 0 {
		tables = []string{
			"account_repos",
			"plc_log_entries",
			"account_infos",
			"pds_repos",
		}
	}
	for _, table := range tables {
		if res := db.Exec("DELETE FROM " + table); res.Error != nil {
			fmt.Printf("clearing table %q: %s", table, res.Error)
		}
	}

	return nil
}

func DropTables(db *gorm.DB) error {
	tables := []string{
		"account_repos",
		"plc_log_entries",
		"account_infos",
		"pds_repos",
	}
	for _, table := range tables {
		if res := db.Exec("DROP TABLE IF EXISTS " + table); res.Error != nil {
			fmt.Printf("dropping table %q: %s", table, res.Error)
		}
	}

	return nil
}
