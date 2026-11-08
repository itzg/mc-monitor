package main

import (
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
)

const (
	BedrockProtocolAuto      = "auto"
	BedrockProtocolNetherNet = "nethernet"
	BedrockProtocolRakNet    = "raknet"
)

// bedrockAutoNetherNetTimeout limits the NetherNet probe in auto mode before falling back to RakNet
const bedrockAutoNetherNetTimeout = 3 * time.Second

// normalizeBedrockProtocol validates the given protocol name, where empty means auto
func normalizeBedrockProtocol(protocol string) (string, error) {
	switch p := strings.ToLower(protocol); p {
	case "", BedrockProtocolAuto:
		return BedrockProtocolAuto, nil
	case BedrockProtocolNetherNet, BedrockProtocolRakNet:
		return p, nil
	default:
		return "", fmt.Errorf("unknown Bedrock protocol '%s', must be auto, nethernet, or raknet", protocol)
	}
}

// PingBedrock queries a Bedrock server with the given protocol. In auto mode, NetherNet is tried
// first and RakNet is used when that fails.
func PingBedrock(address string, protocol string, timeout time.Duration, logger *zap.Logger) (*BedrockServerInfo, error) {
	protocol, err := normalizeBedrockProtocol(protocol)
	if err != nil {
		return nil, err
	}

	switch protocol {
	case BedrockProtocolNetherNet:
		return PingBedrockNetherNet(address, timeout, logger)
	case BedrockProtocolRakNet:
		return PingBedrockServer(address, timeout, logger)
	default:
		netherNetTimeout := bedrockAutoNetherNetTimeout
		if timeout > 0 && timeout < netherNetTimeout {
			netherNetTimeout = timeout
		}
		info, err := PingBedrockNetherNet(address, netherNetTimeout, logger)
		if err != nil {
			logger.Debug("NetherNet probe failed, falling back to RakNet", zap.Error(err))
			return PingBedrockServer(address, timeout, logger)
		}
		return info, nil
	}
}
