package runtime

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"

	"github.com/blebbit/atmunge/pkg/db"
	"gorm.io/gorm/clause"
)

type RepoListResp struct {
	Cursor string `json:"cursor"`
	Repos  []struct {
		Did    string `json:"did"`
		Head   string `json:"head"`
		Rev    string `json:"rev"`
		Active bool   `json:"active"`
		Status string `json:"status"`
	} `json:"repos"`
}

func (r *Runtime) BackfillPdsAccounts(pdses []string, startPDS string) error {

	sort.Strings(pdses)

	skip := startPDS != ""
	for _, pds := range pdses {
		pds = strings.TrimSuffix(pds, "/")

		if skip && pds == startPDS {
			skip = false
		}
		if skip {
			continue
		}

		durl := pds + "/xrpc/com.atproto.server.describeServer"

		r1, err := http.Get(durl)
		if err != nil {
			fmt.Printf("failed to get repos from %s: %w\n", durl, err)
			continue
		}
		if r1 != nil && r1.Body != nil {
			defer r1.Body.Close()
		}

		if r1.StatusCode != http.StatusOK {
			// You might want to read the body here to get more error details
			fmt.Printf("bad status code from %s: %d\n", durl, r1.StatusCode)
			// return fmt.Errorf("bad status code from %s: %d", durl, r1.StatusCode)
			continue
		}

		b1, err := io.ReadAll(r1.Body)
		if err != nil {
			fmt.Printf("failed to read response body from %s: %w\n", durl, err)
			// return fmt.Errorf("failed to read response body from %s: %w", durl, err)
			continue
		}

		fmt.Println(pds+":", string(b1))

		err = r.backfillPdsAccounts(pds)
		if err != nil {
			fmt.Printf("failed while fetching accounts from PDS(%s): %s", pds, err)
			continue
		}

	}

	return nil
}

func (r *Runtime) backfillPdsAccounts(pds string) error {

	cursor := ""
	for {

		// inner loop over accounts
		url := pds + "/xrpc/com.atproto.sync.listRepos?limit=1000"
		if cursor != "" {
			url += "&cursor=" + cursor
		}

		r2, err := http.Get(url)
		if err != nil {
			fmt.Printf("failed to get repos from %s: %w\n", url, err)
			break
		}
		if r2 != nil && r2.Body != nil {
			defer r2.Body.Close()
		}

		if r2.StatusCode != http.StatusOK {
			// You might want to read the body here to get more error details
			fmt.Printf("bad status code from %s: %d\n", url, r2.StatusCode)
			break
		}

		b2, err := io.ReadAll(r2.Body)
		if err != nil {
			fmt.Printf("failed to read response body from %s: %w\n", url, err)
			break
		}

		d2 := RepoListResp{}
		err = json.Unmarshal(b2, &d2)
		if err != nil {
			fmt.Printf("failed to unmarshal response from %s: %w\n", url, err)
			break
		}
		fmt.Printf("Got %d repos from %s @ %s\n", len(d2.Repos), url, d2.Cursor)

		if d2.Cursor == "" {
			fmt.Printf("no cursor found in response from %s\n", url)
			break
		}
		if len(d2.Repos) == 0 {
			fmt.Printf("no repos found at %s\n", url)
			break
		}

		entries := make([]db.PdsRepo, 0, len(d2.Repos))

		// save to the database
		for _, repo := range d2.Repos {
			entries = append(entries, db.PdsRepo{
				PDS:    pds,
				DID:    repo.Did,
				Head:   repo.Head,
				Rev:    repo.Rev,
				Active: repo.Active,
				Status: repo.Status,
			})
		}
		err = r.DB.Table("pds_repos").Clauses(
			clause.OnConflict{
				// TargetWhere: ,
				// Where: ,
				// we only want to update when the head / rev are newer,
				// so that we don't change the updated_at timestamp and
				//   trigger unneccessary work downstream
				Columns: []clause.Column{{Name: "pds"}, {Name: "did"}},
				// we do NOT want to update active or status,
				// as the listRepos always reports active
				// and we have another process which handles these columns
				DoUpdates: clause.AssignmentColumns([]string{"head", "rev"}),
			},
		).Create(entries).Error
		if err != nil {
			fmt.Printf("failed to save repos from %s: %w\n", url, err)
			continue
		}

		// assumes there are no more to fetch
		if len(d2.Repos) < 1000 {
			break
		}

		// update the cursor for the next iteration
		cursor = d2.Cursor
	}

	return nil
}
