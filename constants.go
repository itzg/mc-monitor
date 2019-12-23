package main

const DefaultPort int = 25565

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
