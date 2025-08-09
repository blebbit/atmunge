package runtime

import (
	"fmt"

	"github.com/blebbit/at-mirror/pkg/db"
)

// WARNING: args should not come from user input, this is for internal use only
// likely susceptible to SQL injection
func (r *Runtime) countRemainingToProcess(table string, start, startWhen string) (int, error) {
	q := r.DB.Model(&db.PdsRepo{}).Where("active = true")

	if start == "" {
		q = q.Where(fmt.Sprint("NOT EXISTS (SELECT 1 FROM %s WHERE %s.did = pds_repos.did)", table, table))
	} else {
		q = q.Where(startWhen+" < ?", start)
	}

	var count int64
	err := q.Count(&count).Error
	if err != nil {
		return -1, fmt.Errorf("failed to count PdsRepo entries: %w", err)
	}

	return int(count), nil
}

// WARNING: args should not come from user input, this is for internal use only
// likely susceptible to SQL injection
func (r *Runtime) getRandomSetToProcess(table string, start, startWhen string, limit int) ([]string, error) {
	// fetch all PdsRepo entries that have no corresponding AccountInfo entry
	var ids []string

	q := r.DB.Model(&db.PdsRepo{}).Where("active = true")

	if start == "" {
		q = q.Where(fmt.Sprint("NOT EXISTS (SELECT 1 FROM %s WHERE %s.did = pds_repos.did)", table, table))
	} else {
		q = q.Where(startWhen+" < ?", start)
	}

	q = q.Order("RANDOM()").Limit(limit)

	err := q.Pluck("id", &ids).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch PdsRepo entries: %w", err)
	}

	return ids, nil
}
