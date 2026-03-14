package format

import (
	"encoding/binary"
	"errors"
	"fmt"
)

const (
	Magic           = "ENVC"
	Version    byte = 0x01
	headerSize      = 8
)

type Envelope struct {
	Nonce      []byte
	WrappedKey []byte
	Ciphertext []byte
}

func Marshal(envelope Envelope) ([]byte, error) {
	if len(envelope.Nonce) > 0xFF {
		return nil, errors.New("nonce is too large")
	}
	if len(envelope.WrappedKey) > 0xFFFF {
		return nil, errors.New("wrapped key is too large")
	}

	data := make([]byte, headerSize+len(envelope.Nonce)+len(envelope.WrappedKey)+len(envelope.Ciphertext))
	copy(data[:4], []byte(Magic))
	data[4] = Version
	data[5] = byte(len(envelope.Nonce))
	binary.BigEndian.PutUint16(data[6:8], uint16(len(envelope.WrappedKey)))

	offset := headerSize
	copy(data[offset:offset+len(envelope.Nonce)], envelope.Nonce)
	offset += len(envelope.Nonce)
	copy(data[offset:offset+len(envelope.WrappedKey)], envelope.WrappedKey)
	offset += len(envelope.WrappedKey)
	copy(data[offset:], envelope.Ciphertext)

	return data, nil
}

func Unmarshal(data []byte) (Envelope, error) {
	if len(data) < len(Magic) || string(data[:len(Magic)]) != Magic {
		return Envelope{}, errors.New("invalid file format: missing ENVC header")
	}
	if len(data) < headerSize {
		return Envelope{}, errors.New("corrupted file: unexpected end of data")
	}
	if data[4] != Version {
		return Envelope{}, fmt.Errorf("unsupported file format version: 0x%02x", data[4])
	}

	nonceLen := int(data[5])
	wrappedKeyLen := int(binary.BigEndian.Uint16(data[6:8]))
	offset := headerSize
	requiredSize := offset + nonceLen + wrappedKeyLen
	if len(data) < requiredSize {
		return Envelope{}, errors.New("corrupted file: unexpected end of data")
	}

	return Envelope{
		Nonce:      append([]byte(nil), data[offset:offset+nonceLen]...),
		WrappedKey: append([]byte(nil), data[offset+nonceLen:requiredSize]...),
		Ciphertext: append([]byte(nil), data[requiredSize:]...),
	}, nil
}
