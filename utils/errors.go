package utils

import (
	"fmt"
	"os"
)

func PrintUsageError(msg string) {
	_, _ = fmt.Fprintln(os.Stderr, msg)
}
