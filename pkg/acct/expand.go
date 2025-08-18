package acct

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/blebbit/at-mirror/pkg/repo"
	"github.com/blebbit/at-mirror/pkg/runtime"
	"github.com/rs/zerolog/log"
)

// ExpansionFunc defines the signature for functions that can be used for expansion.
type ExpansionFunc func(rt *runtime.Runtime, did string) error

// expansionFuncs is a map of what-to-expand strings to their corresponding functions.
var expansionFuncs = map[string]ExpansionFunc{
	"ref-records": ExpandRefRecords,
	"ref-repos":   ExpandRefRepos,
}

// GetExpansionKeys returns a slice of strings containing the keys of the expansionFuncs map.
func GetExpansionKeys() []string {
	keys := make([]string, 0, len(expansionFuncs))
	for k := range expansionFuncs {
		keys = append(keys, k)
	}
	return keys
}

func Expand(rt *runtime.Runtime, did string, what string) error {
	if fn, ok := expansionFuncs[what]; ok {
		return fn(rt, did)
	}
	return fmt.Errorf("unknown expansion target: %s", what)
}

func ExpandRefRecords(rt *runtime.Runtime, did string) error {
	log.Info().Str("did", did).Msg("expanding ref repos for account")

	dbPath := filepath.Join(rt.Cfg.RepoDataDir, did, "repo.duckdb")
	db, err := repo.InitDuckDB(dbPath)
	if err != nil {
		// if the duckdb file doesn't exist, that's fine, we just don't have any refs yet
		if os.IsNotExist(err) {
			log.Info().Str("did", did).Msg("no duckdb file found, assuming no refs to expand")
			return nil
		}
		return fmt.Errorf("failed to init duckdb for %s: %w", did, err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT DISTINCT did FROM refs WHERE did IS NOT NULL")
	if err != nil {
		return fmt.Errorf("failed to query refs for %s: %w", did, err)
	}
	defer rows.Close()

	var refDids []string
	for rows.Next() {
		var refDid string
		if err := rows.Scan(&refDid); err != nil {
			return fmt.Errorf("failed to scan ref did for %s: %w", did, err)
		}
		refDids = append(refDids, refDid)
	}

	total := len(refDids)
	log.Info().Str("did", did).Int("total", total).Msg("found ref repos to expand")

	for i, refDid := range refDids {
		log.Info().Str("did", refDid).Int("n", i+1).Int("total", total).Msg("processing ref repo")

		// 1. load duckdb for refDid
		refDbPath := filepath.Join(rt.Cfg.RepoDataDir, refDid, "repo.duckdb")
		refDb, err := repo.InitDuckDB(refDbPath)
		if err != nil {
			if os.IsNotExist(err) {
				log.Warn().Str("did", refDid).Msg("no duckdb file found for ref, skipping")
				continue
			}
			return fmt.Errorf("failed to init duckdb for %s: %w", refDid, err)
		}
		defer refDb.Close()

		// 2. find all refs in did duckdb for refDid
		refRows, err := db.Query("SELECT nsid, rkey FROM refs WHERE did = ? AND record IS NULL", refDid)
		if err != nil {
			return fmt.Errorf("failed to query refs for %s in %s: %w", refDid, did, err)
		}
		defer refRows.Close()

		// 3. lookup record in refDid duckdb and insert into did duckdb
		for refRows.Next() {
			var nsid, rkey string
			if err := refRows.Scan(&nsid, &rkey); err != nil {
				return fmt.Errorf("failed to scan ref for %s in %s: %w", refDid, did, err)
			}

			var recordJSON string
			err := refDb.QueryRow("SELECT record::TEXT FROM records WHERE nsid = ? AND rkey = ?", nsid, rkey).Scan(&recordJSON)
			if err != nil {
				log.Warn().Err(err).Str("did", refDid).Str("nsid", nsid).Str("rkey", rkey).Msg("failed to find record in refDb")
				continue
			}

			_, err = db.Exec(
				"UPDATE refs SET record = ? WHERE did = ? AND nsid = ? AND rkey = ?",
				recordJSON, refDid, nsid, rkey,
			)
			if err != nil {
				return fmt.Errorf("failed to update record from %s into %s: %w", refDid, did, err)
			}
			log.Info().Str("did", did).Str("refDid", refDid).Str("nsid", nsid).Str("rkey", rkey).Msg("updated record from ref")
		}
	}

	return nil
}

func ExpandRefRepos(rt *runtime.Runtime, did string) error {
	log.Info().Str("did", did).Msg("expanding ref repos for account")

	dbPath := filepath.Join(rt.Cfg.RepoDataDir, did, "repo.duckdb")
	db, err := repo.InitDuckDB(dbPath)
	if err != nil {
		// if the duckdb file doesn't exist, that's fine, we just don't have any refs yet
		if os.IsNotExist(err) {
			log.Info().Str("did", did).Msg("no duckdb file found, assuming no refs to expand")
			return nil
		}
		return fmt.Errorf("failed to init duckdb for %s: %w", did, err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT DISTINCT did FROM refs WHERE did IS NOT NULL")
	if err != nil {
		return fmt.Errorf("failed to query refs for %s: %w", did, err)
	}
	defer rows.Close()

	var refDids []string
	for rows.Next() {
		var refDid string
		if err := rows.Scan(&refDid); err != nil {
			return fmt.Errorf("failed to scan ref did for %s: %w", did, err)
		}
		refDids = append(refDids, refDid)
	}

	total := len(refDids)
	log.Info().Str("did", did).Int("total", total).Msg("found ref repos to expand")

	for i, refDid := range refDids {
		log.Info().Str("did", refDid).Int("n", i+1).Int("total", total).Msg("processing ref repo")
		if err := Sync(rt, refDid, []string{"car", "duckdb"}); err != nil {
			log.Error().Err(err).Str("did", refDid).Msg("failed to sync ref repo")
		}
	}

	return nil
}
