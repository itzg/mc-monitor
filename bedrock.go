package main

import (
	"fmt"
	"github.com/sandertv/go-raknet"
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

func PingBedrockServer(address string) (*BedrockServerInfo, error) {
	start := time.Now()
	response, err := raknet.Ping(address)
	rtt := time.Now().Sub(start)
	if err != nil {
		return nil, fmt.Errorf("failed to query bedrock server %s: %w", address, err)
	}

	parts := strings.Split(string(response), ";")
	info := &BedrockServerInfo{
		Rtt:             rtt,
		ServerName:      parts[1],
		ProtocolVersion: parts[2],
		Version:         parts[3],
		Players:         safeParseInt(parts[4]),
		MaxPlayers:      safeParseInt(parts[5]),
		LevelName:       parts[7],
		GameMode:        parts[8],
	}
	if len(parts) >= 10 {
		info.Difficulty = parts[9]
	}

	return info, nil
}

func safeParseInt(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		return -1
	} else {
		return i
	}
}
