package runtime

import (
	"regexp"
	"sort"
)

var KnownBadPDS = []string{
	"https://uwu",           // invalid domain (first spam event)
	"https://pds.trump.com", // spamming, DNS doesn't resolve
}

func init() {
	sort.StringSlice(KnownBadPDS).Sort()
}

var handleRegex = regexp.MustCompile(`^([a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?$`)
