package nstools

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

type CNMT struct {
	Type                          string
	ID                            string `xml:"Id"`
	Version                       int
	RequiredDownloadSystemVersion int
	RequiredSystemVersion         int
	KeyGenerationMin              int
	// TODO: Add more fields
}

// CnmtFromXML tries to parse the XML data and populate the CNMT struct.
// it returns an ErrNotFound error if the cnmt.xml file is not found.
func CnmtFromXML(pfs0 *PFS0) (*CNMT, error) {
	for i, file := range pfs0.Files {
		if strings.HasSuffix("cnmt.xml", file.Name) {
			data := bytes.NewBuffer(nil)
			err := pfs0.Extract(data, i)
			if err != nil {
				return nil, err
			}
			return fromXML(data.Bytes())
		}
	}
	return nil, fmt.Errorf("cnmt.xml %w", ErrNotFound)
}

func CnmtFromNCA(pfs0 *PFS0, keys *Keys) (*CNMT, error) {
	for i, file := range pfs0.Files {
		if strings.HasSuffix("cnmt.nca", file.Name) {
			data := bytes.NewBuffer(nil)
			err := pfs0.Extract(data, i)
			if err != nil {
				return nil, err
			}

			reader := bytes.NewReader(data.Bytes())
			return fromNCA(reader, keys)
		}
	}
	return nil, fmt.Errorf("cnmt.nca %w", ErrNotFound)
}

func fromXML(data []byte) (*CNMT, error) {
	type ContentMeta struct {
		CNMT `xml:"ContentMeta"`
	}
	meta := &ContentMeta{}

	err := xml.Unmarshal(data, meta)

	return &meta.CNMT, err
}

func fromNCA(data io.ReadSeeker, keys *Keys) (*CNMT, error) {
	nca, err := NewNCA(data)
	if err != nil {
		return nil, err
	}
	err = nca.Decrypt(keys)
	if err != nil {
		return nil, err
	}
	fmt.Printf("%+v", nca)
	return nil, nil
}
