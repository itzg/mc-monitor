package mcstatus

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
)

type PacketDataBuilder struct {
	buf bytes.Buffer
}

func (b *PacketDataBuilder) Bytes() []byte {
	return b.buf.Bytes()
}

func (b *PacketDataBuilder) EncodePacketId(id int32) {
	b.EncodeVarInt(id)
}

func (b *PacketDataBuilder) EncodeVarInt(x int32) {
	ux := uint32(x)
	b.EncodeVarUint(ux)
}

func (b *PacketDataBuilder) EncodeVarUint(x uint32) {

	for {
		temp := byte(x & 0b01111111)
		x >>= 7
		if x != 0 {
			temp |= 0b10000000
		}
		b.buf.WriteByte(temp)

		if x == 0 {
			break
		}
	}
}

func (b *PacketDataBuilder) EncodeString(val string) {
	b.EncodeVarInt(int32(len(val)))
	for _, r := range val {
		b.buf.WriteRune(r)
	}
}

func (b *PacketDataBuilder) EncodeUnsignedShort(val uint16) {
	b.buf.WriteByte(byte(val >> 8))
	b.buf.WriteByte(byte(val))
}

func ReadVarInt(r io.Reader) (int32, error) {
	var result uint32

	buf := make([]byte, 1)
	for numRead := 0; numRead <= 5; numRead++ {
		_, err := r.Read(buf)
		if err != nil {
			return 0, err
		}

		value := buf[0] & 0b01111111
		result |= uint32(value) << (7 * numRead)

		if buf[0]&0b10000000 == 0 {
			break
		}
	}

	return int32(result), nil
}

func ReadString(r io.Reader) (string, error) {
	strlen, err := ReadVarInt(r)
	if err != nil {
		return "", err
	}

	buf := make([]byte, strlen)
	_, err = io.ReadFull(r, buf)
	if err != nil {
		return "", err
	}

	return string(buf), nil
}

type State int32

const (
	StateStatus State = 1
	StateLogin  State = 2
)

type ServerBoundPacket interface {
	Encode() []byte
}

type Handshake struct {
	ProtocolVersion int32
	ServerAddress   string
	ServerPort      uint16
	NextState       State
}

func (h *Handshake) Encode() []byte {
	builder := &PacketDataBuilder{}

	builder.EncodePacketId(0)

	builder.EncodeVarInt(h.ProtocolVersion)
	builder.EncodeString(h.ServerAddress)
	builder.EncodeUnsignedShort(h.ServerPort)
	builder.EncodeVarInt(int32(h.NextState))

	return builder.Bytes()
}

type StateRequest struct{}

func (s *StateRequest) Encode() []byte {
	builder := &PacketDataBuilder{}
	builder.EncodePacketId(0)
	return builder.Bytes()
}

type ClientBoundPacket interface {
	Decode(r io.Reader) error
}

type StateResponse struct {
	Version struct {
		Name     string
		Protocol int
	}
	Players struct {
		Max    int
		Online int
	}
	Description struct {
		Text string
	}
	Favicon string
}

func (s *StateResponse) Decode(r io.Reader) error {
	jsonContent, err := ReadString(r)
	if err != nil {
		return err
	}

	err = json.Unmarshal([]byte(jsonContent), s)
	if err != nil {
		return err
	}

	return nil
}

type Client struct {
	address string
	port    uint16
	conn    net.Conn
	state   State
}

func NewClient(address string, port uint16) *Client {
	return &Client{address: address, port: port}
}

func (c *Client) Connect() error {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", c.address, c.port))
	if err != nil {
		return err
	}
	c.conn = conn
	return nil
}

func (c *Client) Handshake() (*StateResponse, error) {
	err := c.Send(&Handshake{
		ProtocolVersion: -1,
		ServerAddress:   c.address,
		ServerPort:      c.port,
		NextState:       StateStatus,
	})
	if err != nil {
		return nil, err
	}

	c.state = StateStatus

	err = c.Send(&StateRequest{})
	if err != nil {
		return nil, err
	}

	frameReader, err := c.waitForFrame()
	if err != nil {
		return nil, fmt.Errorf("failed to get response frame: %w", err)
	}

	packetId, err := ReadVarInt(frameReader)
	if err != nil {
		return nil, err
	}

	if packetId != 0 {
		return nil, fmt.Errorf("invalid packet id: %d", packetId)
	}

	resp := &StateResponse{}
	err = resp.Decode(frameReader)
	if err != nil {
		return nil, fmt.Errorf("failed to decode state response: %w", err)
	}

	return resp, nil
}

func (c *Client) Send(packet ServerBoundPacket) error {
	packetBytes := packet.Encode()

	lenBuilder := &PacketDataBuilder{}
	lenBuilder.EncodeVarInt(int32(len(packetBytes)))

	_, err := c.conn.Write(lenBuilder.Bytes())
	if err != nil {
		return err
	}

	_, err = c.conn.Write(packetBytes)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) waitForFrame() (packetReader io.Reader, err error) {
	var packetLen int32
	packetLen, err = ReadVarInt(c.conn)
	if err != nil {
		return
	}

	packetReader = io.LimitReader(c.conn, int64(packetLen))
	return
}
