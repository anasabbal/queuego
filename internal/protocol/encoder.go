package protocol

import (
	"bytes"
	"encoding/binary"
	"errors"
)

func Encode(cmd *Command) ([]byte, error) {
	if cmd.Type == "" {
		return nil, errors.New("command type is required")
	}
	if cmd.Payload == nil {
		cmd.Payload = []byte{}
	}

	var buf bytes.Buffer

	// 1. Command type (1 byte)
	buf.WriteByte(commandTypeToByte(cmd.Type))

	// 2. Topic  (uint16 length + bytes)
	topicBytes := []byte(cmd.Topic)
	if len(topicBytes) > 65535 {
		return nil, errors.New("topic too long")
	}
	if err := binary.Write(&buf, binary.BigEndian, uint16(len(topicBytes))); err != nil {
		return nil, err
	}
	buf.Write(topicBytes)

	// 3. Payload (uint32 length + bytes)
	if err := binary.Write(&buf, binary.BigEndian, uint32(len(cmd.Payload))); err != nil {
		return nil, err
	}
	buf.Write(cmd.Payload)

	return buf.Bytes(), nil
}

// commandTypeToByte maps CommandType to a single byte
func commandTypeToByte(t CommandType) byte {
	switch t {
	case CONNECT:
		return 0x01
	case PUBLISH:
		return 0x02
	case SUBSCRIBE:
		return 0x03
	case UNSUBSCRIBE:
		return 0x04
	case ACK:
		return 0x05
	case PING:
		return 0x06
	case PONG:
		return 0x07
	default:
		return 0x00
	}
}
