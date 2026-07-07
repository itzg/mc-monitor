package main

import (
	"fmt"

	"github.com/sandertv/go-raknet"
	"go.uber.org/zap"

	"strconv"
	"strings"
	"time"
)

type BedrockServerInfo struct {
	ServerName      string
	ProtocolVersion string
	Version         string
	Players         int
	MaxPlayers      int
	LevelName       string
	GameMode        string
	Difficulty      string
	Rtt             time.Duration
}

func PingBedrockServer(address string, timeout time.Duration, logger *zap.Logger) (*BedrockServerInfo, error) {
	start := time.Now()
	var response []byte
	var err error
	if timeout > 0 {
		response, err = raknet.PingTimeout(address, timeout)
	} else {
		response, err = raknet.Ping(address)
	}
	rtt := time.Now().Sub(start)
	if err != nil {
		return nil, fmt.Errorf("failed to query bedrock server %s: %w", address, err)
	}

	logger.Debug("received response from bedrock server", zap.String("address", address), zap.ByteString("response", response))
	parts := strings.Split(string(response), ";")
	info := &BedrockServerInfo{
		Rtt:             rtt,
		ServerName:      parts[1],
		ProtocolVersion: parts[2],
		Version:         parts[3],
		// the parts past here are not always present in the response
		Players:    safeIntAt(parts, 4),
		MaxPlayers: safeIntAt(parts, 5),
		LevelName:  safeStringAt(parts, 7),
		GameMode:   safeStringAt(parts, 8),
		Difficulty: safeStringAt(parts, 9),
	}

	return info, nil
}

func safeStringAt(parts []string, index int) string {
	if index >= 0 && index < len(parts) {
		return parts[index]
	}
	return ""
}

func safeIntAt(parts []string, index int) int {
	if index >= 0 && index < len(parts) {
		return safeParseInt(parts[index])
	}
	return -1
}

func safeParseInt(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		return -1
	} else {
		return i
	}
}
