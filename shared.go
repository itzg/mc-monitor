package main

import (
	"fmt"
	"os"
)

const (
	// @deprecated use utils.DefaultJavaPort instead
	DefaultJavaPort uint16 = 25565
	// @deprecated use utils.DefaultBedrockPort instead
	DefaultBedrockPort uint16 = 19132
)

// @deprecated use utils.ServerEdition instead
type ServerEdition string

const (
	// @deprecated use utils.JavaEdition instead
	JavaEdition ServerEdition = "java"
	// @deprecated use utils.BedrockEdition instead
	BedrockEdition ServerEdition = "bedrock"
)

// @deprecated use utils.ServerEdition instead
func ValidEdition(v string) bool {
	switch ServerEdition(v) {
	case JavaEdition, BedrockEdition:
		return true
	}
	return false
}

// @deprecated use utils.PrintUsageError instead
func printUsageError(msg string) {
	_, _ = fmt.Fprintln(os.Stderr, msg)
}
