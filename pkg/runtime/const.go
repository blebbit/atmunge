package runtime

import (
	"regexp"
	"sort"
	"time"

	"golang.org/x/time/rate"
)

var KnownBadPDS = []string{
	"https://uwu",           // invalid domain (first spam event)
	"https://pds.trump.com", // spamming, DNS doesn't resolve
}

func init() {
	sort.StringSlice(KnownBadPDS).Sort()
}

var handleRegex = regexp.MustCompile(`^([a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?$`)

const (
	// plc.directory settings
	// default is 500;300w ... aim slightly below that
	plcRateLimit = rate.Limit(480.0 / 300.0)
	plcMaxDelay  = 5 * time.Minute

	// PDS settings (assume consistent, can store exceptions in the PDS info table)
	// default is 3000;300w ... aim slightly below that
	pdsRateLimit = rate.Limit(2900.0 / 300.0)
)
