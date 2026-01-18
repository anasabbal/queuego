package protovol

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

func Decode(data []byte) (*Command, error) {
	if len(data) < 8 {
		return nil, fmt.Errorf("data too short to decode: %d bytes", len(data))
	}

	buf := bytes.NewReader(data)

	readString := func() (string, error) {
		var length uint16
		if err := binary.Read(buf, binary.BigEndian, &length); err != nil {
			return "", fmt.Errorf("failed to read string length: %w", err)
		}

		if length == 0 {
			return "", nil
		}

		strBytes := make([]byte, length)
		n, err := buf.Read(strBytes)
		if err != nil {
			return "", fmt.Errorf("failed to read string bytes: %w", err)
		}
		if n != int(length) {
			return "", fmt.Errorf("string length mismatch: expected %d, got %d", length, n)
		}

		return string(strBytes), nil
	}

	// read Type
	cmdTypeStr, err := readString()
	if err != nil {
		return nil, fmt.Errorf("invalid command type: %w", err)
	}
	if cmdTypeStr == "" {
		return nil, fmt.Errorf("command type is empty")
	}

	// read Topic
	topic, err := readString()
	if err != nil {
		return nil, fmt.Errorf("invalid topic: %w", err)
	}

	// read MessageID
	messageID, err := readString()
	if err != nil {
		return nil, fmt.Errorf("invalid message ID: %w", err)
	}

	// read Payload
	var payloadLen uint32
	if err := binary.Read(buf, binary.BigEndian, &payloadLen); err != nil {
		return nil, fmt.Errorf("failed to read payload length: %w", err)
	}

	if int(payloadLen) != buf.Len() {
		return nil, fmt.Errorf("payload length mismatch: expected %d, remaining %d", payloadLen, buf.Len())
	}

	payload := make([]byte, payloadLen)
	if _, err := buf.Read(payload); err != nil {
		return nil, fmt.Errorf("failed to read payload: %w", err)
	}

	return &Command{
		Type:      CommandType(cmdTypeStr),
		Topic:     topic,
		MessageID: messageID,
		Payload:   payload,
	}, nil
}
