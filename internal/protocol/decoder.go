package protocol

import (
	"bytes"
	"encoding/binary"
	"errors"
)

func Decode(data []byte) (*Command, error) {
	if len(data) < 3 { // min 1 byte type + 2 bytes topic length
		return nil, errors.New("data too short to decode")
	}

	buf := bytes.NewReader(data)

	// 1. Command type
	cmdTypeByte, err := buf.ReadByte()
	if err != nil {
		return nil, err
	}
	cmdType := byteToCommandType(cmdTypeByte)
	if cmdType == "" {
		return nil, errors.New("invalid command type")
	}

	// 2. Topic
	var topicLen uint16
	if err := binary.Read(buf, binary.BigEndian, &topicLen); err != nil {
		return nil, err
	}
	topicBytes := make([]byte, topicLen)
	if _, err := buf.Read(topicBytes); err != nil {
		return nil, err
	}
	topic := string(topicBytes)

	// 3. Payload
	var payloadLen uint32
	if err := binary.Read(buf, binary.BigEndian, &payloadLen); err != nil {
		return nil, err
	}
	payload := make([]byte, payloadLen)
	if _, err := buf.Read(payload); err != nil {
		return nil, err
	}

	return &Command{
		Type:    cmdType,
		Topic:   topic,
		Payload: payload,
	}, nil
}

func byteToCommandType(b byte) CommandType {
	switch b {
	case 0x01:
		return CONNECT
	case 0x02:
		return PUBLISH
	case 0x03:
		return SUBSCRIBE
	case 0x04:
		return UNSUBSCRIBE
	case 0x05:
		return ACK
	case 0x06:
		return PING
	case 0x07:
		return PONG
	default:
		return ""
	}
}
