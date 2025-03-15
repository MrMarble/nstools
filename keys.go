package nstools

import (
	"bytes"
	"encoding/hex"
	"fmt"
)

type Keys struct {
	HeaderKey []byte
}

// NewKeys creates a new Keys struct from the given data.
// The data has the following format:
//
// key_name = hex_key_value
//
// key_name2 = hex_key_value2
func NewKeys(data []byte) *Keys {
	keys := &Keys{}
	lines := bytes.SplitSeq(data, []byte("\n"))
	for line := range lines {
		parts := bytes.Split(line, []byte("="))
		if len(parts) != 2 {
			continue
		}
		key := bytes.TrimSpace(parts[0])
		value := bytes.TrimSpace(parts[1])
		switch string(key) {
		case "header_key":
			key, err := hex.DecodeString(string(value))
			if err != nil {
				continue
			}
			keys.HeaderKey = key
		}
	}
	return keys
}

// Validate checks if the keys are valid.
func (keys *Keys) Validate() error {
	if keys.HeaderKey == nil || len(keys.HeaderKey) != 0x20 {
		return fmt.Errorf("invalid header key %w", ErrInvalidKey)
	}
	return nil
}
