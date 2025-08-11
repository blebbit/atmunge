package util

import (
	"io"
	"os"
)

// GetInput reads from stdin if the arg is "-", otherwise returns the arg.
func GetInput(arg string) (string, error) {
	if arg == "-" {
		bytes, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", err
		}
		return string(bytes), nil
	}
	return arg, nil
}
