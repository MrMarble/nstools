// Package nstools provides tools for working with Nintendo Switch files.
package nstools

import (
	"errors"
	"os"
	"path/filepath"
)

// Errors
var (
	ErrInvalidMagic = errors.New("invalid magic signature")
	ErrUnknownFile  = errors.New("unknown file type")
	ErrInvalidSig   = errors.New("invalid signature type")
	ErrNotFound     = errors.New("file not found")
	ErrInvalidKey   = errors.New("invalid key")
)

// Magic signatures
var (
	MagicPFS0 = [...]byte{0x50, 0x46, 0x53, 0x30} // PFS0
	MagicNCA2 = [...]byte{0x4E, 0x43, 0x41, 0x32} // NCA2
	MagicNCA3 = [...]byte{0x4E, 0x43, 0x41, 0x33} // NCA3
)

//go:generate go tool stringer --type=ContentType --trimprefix=ContentType
type ContentType uint8

const (
	ContentTypeProgram ContentType = iota
	ContentTypeMeta
	ContentTypeControl
	ContentTypeManual
	ContentTypeData
	ContentTypePublicData
)

//go:generate go tool stringer --type=DistributionType --trimprefix=DistributionType
type DistributionType uint8

const (
	DistributionTypeDownload DistributionType = iota
	DistributionTypeGameCard
)

func OpenFile(path string, keys *Keys) (any, error) {
	ext := filepath.Ext(path)

	switch ext {
	case ".nsp", ".pfs0":
		f, err := os.OpenFile(path, os.O_RDONLY, 0644)
		if err != nil {
			return nil, err
		}
		return OpenPFS0(f)
	case ".nca":
		f, err := os.OpenFile(path, os.O_RDONLY, 0644)
		if err != nil {
			return nil, err
		}
		return fromNCA(f, keys)
	default:
		return nil, ErrUnknownFile
	}
}
