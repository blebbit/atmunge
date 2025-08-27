package runtime

import (
	"net/url"
	"strings"

	"github.com/blebbit/atmunge/pkg/plc"
)

func plcUrl() *url.URL {
	u, err := url.Parse("https://plc.directory")
	if err != nil {
		panic(err)
	}
	u.Path, err = url.JoinPath(u.Path, "export")
	if err != nil {
		panic(err)
	}
	return u
}

func validateOperation(entry plc.OperationLogEntry, op plc.Op) bool {

	// check did
	if entry.DID == "" {
		return false
	}

	// check handle
	if len(op.AlsoKnownAs) > 0 {
		handle := strings.TrimPrefix(op.AlsoKnownAs[0], "at://")
		// check if handle is a valid handle
		if _, err := url.Parse(handle); err != nil {
			return false
		}
	}

	// check pds
	if svc, ok := op.Services["atproto_pds"]; ok {
		pds := svc.Endpoint
		if _, err := url.Parse(pds); err != nil {
			return false
		}

		if pds == "https://uwu" {
			return false
		}

		// check some well-known bad values
		switch pds {

		}
	}

	return true
}
