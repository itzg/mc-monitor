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

func PingBedrockServer(address string, timeout time.Duration) (*BedrockServerInfo, error) {
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

	parts := strings.Split(string(response), ";")
	info := &BedrockServerInfo{
		Rtt:             rtt,
		ServerName:      parts[1],
		ProtocolVersion: parts[2],
		Version:         parts[3],
		Players:         safeParseInt(parts[4]),
		MaxPlayers:      safeParseInt(parts[5]),
	}
	if len(parts) > 7 {
		info.LevelName = parts[7]
		info.GameMode = parts[8]
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
