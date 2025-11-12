// Package slp implements Server List Ping and Old Server List Ping accepted by servers to version 1.3
// before 1.6; however, modern servers also respond to it.
package slp

import (
	"encoding/binary"
	"fmt"
	"strings"
	"io"
	"net"
	"strconv"
	"time"
)

type OldServerListResponse struct {
	MessageOfTheDay    string
	CurrentPlayerCount string
	MaxPlayers         string
}

func OldServerListPing(host string, port int, timeout time.Duration) (*OldServerListResponse, error) {
	conn, err := net.Dial("tcp", net.JoinHostPort(host, strconv.Itoa(port)))
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}
	//goland:noinspection GoUnhandledErrorResult
	defer conn.Close()

	err = encodePingOld(conn, host, port)
	if err != nil {
		return nil, fmt.Errorf("failed to send ping: %w", err)
	}

	var packetId = make([]byte, 1)
	_ = conn.SetReadDeadline(time.Now().Add(timeout))
	_, err = conn.Read(packetId)
	if err != nil {
		return nil, fmt.Errorf("failed to read response packet ID: %w", err)
	}

	if packetId[0] != 0xff {
		return nil, fmt.Errorf("invalid packet ID received from server: %x", packetId[0])
	}

	var contentLen uint16
	err = binary.Read(conn, binary.BigEndian, &contentLen)
	if err != nil {
		return nil, fmt.Errorf("failed to read content length: %w", err)
	}

	serverResponse, err := readString(conn)
	if err != nil {
		return nil, fmt.Errorf("failed to read server response: %w", err)
	}

	responseSplit := strings.Split(serverResponse, "§")
	messageOfTheDay := responseSplit[0]
	currentPlayerCount := responseSplit[1]
	maxPlayers := responseSplit[2]
	
	var response OldServerListResponse
	response.MessageOfTheDay = messageOfTheDay
	response.CurrentPlayerCount = currentPlayerCount
	response.MaxPlayers = maxPlayers

	return &response, nil
}

func encodePingOld(conn io.Writer, host string, port int) error {
	//see https://c4k3.github.io/wiki.vg/Server_List_Ping.html#Beta_1.8_to_1.3
	err := writeBinarySlice(conn, []interface{}{
		uint8(0xFE),
	})
	if err != nil {
		return fmt.Errorf("failed to encode server list ping: %w", err)
	}
	return nil
}
