// Package slp implements the Server List Ping protocol originally accepted by servers
// before 1.6; however, modern servers also respond to it.
package slp

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strconv"
	"time"
	"unicode/utf16"
)

type ServerListResponse struct {
	ProtocolVersion    string
	ServerVersion      string
	MessageOfTheDay    string
	CurrentPlayerCount string
	MaxPlayers         string
}

func ServerListPing(host string, port int, timeout time.Duration) (*ServerListResponse, error) {
	conn, err := net.Dial("tcp", net.JoinHostPort(host, strconv.Itoa(port)))
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}
	//goland:noinspection GoUnhandledErrorResult
	defer conn.Close()

	err = encodePing(conn, host, port)
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

	header, err := readString(conn)
	if err != nil {
		return nil, fmt.Errorf("failed to read header: %w", err)
	}
	if header != "§1" {
		return nil, fmt.Errorf("invalid response header: %s", header)
	}

	protocolVersion, err := readString(conn)
	if err != nil {
		return nil, fmt.Errorf("failed to read protocolVersion: %w", err)
	}
	serverVersion, err := readString(conn)
	if err != nil {
		return nil, fmt.Errorf("failed to read serverVersion: %w", err)
	}
	messageOfTheDay, err := readString(conn)
	if err != nil {
		return nil, fmt.Errorf("failed to read messageOfTheDay: %w", err)
	}
	currentPlayerCount, err := readString(conn)
	if err != nil {
		return nil, fmt.Errorf("failed to read currentPlayerCount: %w", err)
	}
	maxPlayers, err := readString(conn)
	if err != nil {
		return nil, fmt.Errorf("failed to read maxPlayers: %w", err)
	}

	var response ServerListResponse
	response.ProtocolVersion = protocolVersion
	response.ServerVersion = serverVersion
	response.MessageOfTheDay = messageOfTheDay
	response.CurrentPlayerCount = currentPlayerCount
	response.MaxPlayers = maxPlayers

	return &response, nil
}

// readString reads a null terminated UTF-16BE string from the reader
func readString(conn io.Reader) (string, error) {
	chars := make([]uint16, 0)

	var next uint16
	for {
		err := binary.Read(conn, binary.BigEndian, &next)
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", err
		} else if next == 0 {
			break
		} else {
			chars = append(chars, next)
		}
	}

	return string(utf16.Decode(chars)), nil
}

func encodePing(conn io.Writer, host string, port int) error {
	// see https://wiki.vg/Server_List_Ping#Client_to_server
	err := writeBinarySlice(conn, []interface{}{
		uint8(0xFE),
		uint8(1),
		uint8(0xFA),
		uint16(11),
	})
	if err != nil {
		return fmt.Errorf("failed to encode server list ping: %w", err)
	}
	err = writeBinaryUtf16(conn, utf16.Encode([]rune("MC|PingHost")))
	if err != nil {
		return fmt.Errorf("failed to encode server list ping: %w", err)
	}
	hostEncoded := new(bytes.Buffer)
	_ = writeBinaryUtf16(hostEncoded, utf16.Encode([]rune(host)))
	err = writeBinarySlice(conn, []interface{}{
		uint16(7 + hostEncoded.Len()),
		uint8(74),
		// length of following string, in characters, as a short
		uint16(len(host)),
	})
	if err != nil {
		return fmt.Errorf("failed to encode server list ping: %w", err)
	}
	_, err = io.Copy(conn, hostEncoded)
	if err != nil {
		return fmt.Errorf("failed to write host: %w", err)
	}
	err = binary.Write(conn, binary.BigEndian, uint32(port))
	if err != nil {
		return fmt.Errorf("failed to encode server list ping: %w", err)
	}

	return nil
}

func writeBinaryUtf16(dst io.Writer, encoded []uint16) error {
	for _, v := range encoded {
		err := binary.Write(dst, binary.BigEndian, v)
		if err != nil {
			return err
		}
	}
	return nil
}

func writeBinarySlice(dst io.Writer, data []interface{}) error {
	for _, v := range data {
		err := binary.Write(dst, binary.BigEndian, v)
		if err != nil {
			return err
		}
	}
	return nil
}
