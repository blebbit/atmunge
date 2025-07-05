package runtime

import (
	"net/url"
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
