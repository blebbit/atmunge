package runtime

import (
	"fmt"

	atdb "github.com/blebbit/at-mirror/pkg/db"
)

// WARNING: table name should not come from user input, this is for internal use only
func (r *Runtime) countRemainingToProcess(table string, start string) (int, error) {
	q := r.DB.Model(&atdb.PdsRepo{}).Where("active = true")

	if start == "" {
		q = q.Where(fmt.Sprint("NOT EXISTS (SELECT 1 FROM %s WHERE %s.did = pds_repos.did)", table, table))
	} else {
		q = q.Where("updated_at < ?", start)
	}

	var count int64
	err := q.Count(&count).Error
	if err != nil {
		return -1, fmt.Errorf("failed to count PdsRepo entries: %w", err)
	}

	return int(count), nil
}

// WARNING: table name should not come from user input, this is for internal use only
func (r *Runtime) getRandomSetToProcess(table string, start string, limit int) ([]string, error) {
	// fetch all PdsRepo entries that have no corresponding AccountInfo entry
	var ids []string

	q := r.DB.Model(&atdb.PdsRepo{}).Where("active = true")

	if start == "" {
		q = q.Where(fmt.Sprint("NOT EXISTS (SELECT 1 FROM %s WHERE %s.did = pds_repos.did)", table, table))
	} else {
		q = q.Where("updated_at < ?", start)
	}

	q = q.Order("RANDOM()").Limit(limit)

	err := q.Pluck("id", &ids).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch PdsRepo entries: %w", err)
	}

	return ids, nil
}
