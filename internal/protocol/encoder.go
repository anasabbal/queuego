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
	var buf bytes.Buffer

	// 1 - command type (1 byte)
	buf.WriteByte(commandTypeToByte(cmd.Type))

	// 2 - Topic
	topicBytes := []byte(cmd.Topic)
	if len(topicBytes) > 65535 {
		return nil, errors.New("topic too long")
	}
	if err := binary.Write(&buf, binary.BigEndian, uint16(len(topicBytes))); err != nil {
		return nil, err
	}

	// 3 - Payload
	payload := cmd.Payload
	if payload == nil {
		payload = []byte{}
	}
	if err := binary.Write(&buf, binary.BigEndian, uint32(len(payload))); err != nil {
		return nil, err
	}
	if len(payload) > 0 {
		buf.Write(payload)
	}

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
