package main

import (
	"fmt"
	"os"
)

const (
	DefaultJavaPort    uint16 = 25565
	DefaultBedrockPort uint16 = 19132
)

type ServerEdition string

const (
	JavaEdition    ServerEdition = "java"
	BedrockEdition ServerEdition = "bedrock"
)

func ValidEdition(v string) bool {
	switch ServerEdition(v) {
	case JavaEdition, BedrockEdition:
		return true
	}
	return false
}

func printUsageError(msg string) {
	_, _ = fmt.Fprintln(os.Stderr, msg)
}
