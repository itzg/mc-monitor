package main

import (
	"fmt"
	"strconv"
	"strings"
)

func SplitHostPort(hostport string, defaultPort uint16) (string, uint16, error) {
	parts := strings.SplitN(hostport, ":", 2)
	if len(parts) == 2 {
		parsed, err := strconv.ParseUint(parts[1], 10, 16)
		if err != nil {
			return "", 0, err
		}
		return parts[0], uint16(parsed), nil
	} else {
		return parts[0], defaultPort, nil
	}
}

func NormalizeHostPort(hostport string, defaultPort uint16) string {
	if strings.Contains(hostport, ":") {
		return hostport
	} else {
		return fmt.Sprintf("%s:%d", hostport, defaultPort)
	}
}
