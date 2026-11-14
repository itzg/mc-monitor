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

func PingBedrockServer(address string, nethernetMode string, protocol string, timeout time.Duration, logger *zap.Logger) (*BedrockServerInfo, error) {
	switch strings.ToLower(protocol) {
	case "nethernet":
		return PingBedrockNetherNet(address, nethernetMode, timeout, logger)
	case "raknet":
		return PingBedrockRaknetServer(address, timeout, logger)
	case "auto", "":
		info, err := PingBedrockNetherNet(address, nethernetMode, 3*time.Second, logger)
		if err != nil {
			logger.Debug("NetherNet probe failed, falling back to RakNet", zap.Error(err))
			return PingBedrockRaknetServer(address, timeout, logger)
		}
		return info, nil
	default:
		logger.Fatal("Unknown protocol", zap.String("protocol", protocol))
		return nil, fmt.Errorf("unknown protocol: %s", protocol)
	}
}

func PingBedrockRaknetServer(address string, timeout time.Duration, logger *zap.Logger) (*BedrockServerInfo, error) {
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
	if len(response) == 0 {
		return nil, fmt.Errorf("empty response from bedrock server %s", address)
	}

	parts := strings.Split(string(response), ";")
	info := &BedrockServerInfo{
		Rtt:             rtt,
		ServerName:      safeStringAt(parts, 1),
		ProtocolVersion: safeStringAt(parts, 2),
		Version:         safeStringAt(parts, 3),
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
