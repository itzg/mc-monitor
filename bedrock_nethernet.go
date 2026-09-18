package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
)

type netherNetResponse struct {
	Name       string      `json:"name"`
	Protocol   interface{} `json:"protocol"`
	Version    string      `json:"version"`
	Level      string      `json:"level"`
	Players    int         `json:"players"`
	MaxPlayers int         `json:"maxPlayers"`
	GameType   int         `json:"gameType"`
}

func PingBedrockNetherNet(address string, timeout time.Duration, logger *zap.Logger) (*BedrockServerInfo, error) {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	client := &http.Client{
		Timeout: timeout,
	}

	url := address
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "http://" + url
	}
	if !strings.HasSuffix(url, "/v1/join") {
		url = strings.TrimSuffix(url, "/") + "/v1/join"
	}

	logger.Debug("Querying bedrock server using NetherNet", zap.String("url", url))
	start := time.Now()
	resp, err := client.Get(url)
	rtt := time.Since(start)
	if err != nil {
		return nil, fmt.Errorf("Failed to query bedrock server %s: %w", address, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Bedrock server %s returned status %d", address, resp.StatusCode)
	}

	var data netherNetResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("Failed to decode response from bedrock server %s: %w", address, err)
	}

	logger.Debug("Received response from bedrock server using NetherNet", zap.String("address", address), zap.Any("response", data))

	info := &BedrockServerInfo{
		Rtt:             rtt,
		ServerName:      data.Name,
		ProtocolVersion: fmt.Sprintf("%v", data.Protocol),
		Version:         data.Version,
		Players:         data.Players,
		MaxPlayers:      data.MaxPlayers,
		LevelName:       data.Level,
		GameMode:        gameTypeToString(data.GameType),
	}

	return info, nil
}

func gameTypeToString(gt int) string {
	switch gt {
	case 0:
		return "Survival"
	case 1:
		return "Creative"
	case 2:
		return "Adventure"
	case 5:
		return "Default"
	case 6:
		return "Spectator"
	default:
		return strconv.Itoa(gt)
	}
}
