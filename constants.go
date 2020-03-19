package main

const (
	DefaultJavaPort    uint16 = 25565
	DefaultBedrockPort uint16 = 19132
)

const (
	MetricName = "minecraft_status"

	TagHost   = "host"
	TagPort   = "port"
	TagStatus = "status"

	FieldError        = "error"
	FieldOnline       = "online"
	FieldMax          = "max"
	FieldResponseTime = "response_time"

	StatusError   = "error"
	StatusSuccess = "success"
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
