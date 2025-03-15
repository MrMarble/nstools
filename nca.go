package nstools

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/mrmarble/nstools/internal/aesxts"
)

type FSEntry struct {
	StartOffset uint32
	EndOffset   uint32
}

type NCAHeader struct {
	FirstSignature   [0x100]byte
	SecondSignature  [0x100]byte
	Magic            [4]byte
	DistributionType DistributionType
	ContentType      ContentType
	KeyGenerationOld uint8
	KeyIndex         uint8
	ContentSize      uint64
	ProgramID        string
	ContentIndex     uint32
	SDKVersion       uint32 // {byte3}.{byte2}.{byte1}, byte0 is always 0
	KeyGeneration    uint32
	RigthsID         string
	SectionTables    [4]FSEntry
	SectionHashes    [4][0x20]byte
}

type NCA struct {
	data io.ReadSeeker
	NCAHeader
}

func NewNCA(data io.ReadSeeker) (*NCA, error) {
	return &NCA{data: data}, nil
}

func (nca *NCA) readHeader(keys *Keys) error {
	if keys == nil || keys.HeaderKey == nil {
		return ErrInvalidKey
	}

	encHeader := make([]byte, 0xC00)
	binary.Read(nca.data, binary.LittleEndian, &encHeader)
	aesCtx, err := aesxts.NewAESCtx(keys.HeaderKey[:])
	if err != nil {
		return err
	}
	defer aesxts.FreeAESCtx(aesCtx)

	header := make([]byte, 0xC00)
	err = aesxts.AESXTSDecrypt(aesCtx, header, encHeader, 0xC00, 0, 0x200)
	if err != nil {
		return err
	}

	programID := make([]byte, 0x08)
	rightsID := make([]byte, 0x10)

	buf := bytes.NewReader(header)
	binary.Read(buf, binary.LittleEndian, &nca.FirstSignature)
	binary.Read(buf, binary.LittleEndian, &nca.SecondSignature)
	binary.Read(buf, binary.LittleEndian, &nca.Magic)

	// No need to continue if the magic is invalid
	if nca.Magic != MagicNCA2 && nca.Magic != MagicNCA3 {
		return fmt.Errorf("invalid NCA magic: %s", nca.Magic)
	}

	binary.Read(buf, binary.LittleEndian, &nca.DistributionType)
	binary.Read(buf, binary.LittleEndian, &nca.ContentType)
	binary.Read(buf, binary.LittleEndian, &nca.KeyGenerationOld)
	binary.Read(buf, binary.LittleEndian, &nca.KeyIndex)
	binary.Read(buf, binary.LittleEndian, &nca.ContentSize)
	binary.Read(buf, binary.LittleEndian, &programID)
	binary.Read(buf, binary.LittleEndian, &nca.ProgramID)
	binary.Read(buf, binary.LittleEndian, &nca.ContentIndex)
	binary.Read(buf, binary.LittleEndian, &nca.SDKVersion)
	binary.Read(buf, binary.LittleEndian, &nca.KeyGeneration)
	buf.Seek(0xF, io.SeekCurrent) // Padding
	binary.Read(buf, binary.LittleEndian, &rightsID)

	nca.RigthsID = hex.EncodeToString(rightsID)
	slices.Reverse(programID)
	nca.ProgramID = hex.EncodeToString(programID)

	for i := range nca.SectionTables {
		binary.Read(buf, binary.LittleEndian, &nca.SectionTables[i].StartOffset)
		binary.Read(buf, binary.LittleEndian, &nca.SectionTables[i].EndOffset)
		buf.Seek(0x08, io.SeekCurrent) // Padding
	}

	for i := range nca.SectionHashes {
		binary.Read(buf, binary.LittleEndian, &nca.SectionHashes[i])
	}

	return nil
}

func (nca *NCA) Decrypt(keys *Keys) error {
	if err := nca.readHeader(keys); err != nil {
		return err
	}
	return nil
}

func (ncah *NCAHeader) String() string {
	t := table.NewWriter()
	t.Style().Box = table.BoxStyle{
		MiddleVertical: "\t",
	}
	t.AppendRow(table.Row{"NCA"})
	t.AppendRow(table.Row{"Magic", string(ncah.Magic[:])})
	t.AppendRow(table.Row{"Fixed-Key Index", ncah.KeyIndex})
	t.AppendRow(table.Row{"Fixed-Key Signature", text.WrapHard(strings.ToUpper(hex.EncodeToString(ncah.FirstSignature[:])), 64)})
	t.AppendRow(table.Row{"NPDM Signature", text.WrapHard(strings.ToUpper(hex.EncodeToString(ncah.SecondSignature[:])), 64)})
	t.AppendRow(table.Row{"Content Size", ncah.ContentSize})
	t.AppendRow(table.Row{"Title ID", ncah.ProgramID})
	t.AppendRow(table.Row{"SDK Version", fmt.Sprintf("%d.%d.%d.%d", byte(ncah.SDKVersion>>24), byte(ncah.SDKVersion>>16), byte(ncah.SDKVersion>>8), byte(ncah.SDKVersion))})
	t.AppendRow(table.Row{"Distribution type", ncah.DistributionType})
	t.AppendRow(table.Row{"Content type", ncah.ContentType})

	return t.Render()
}

func (nca *NCA) String() string {
	return nca.NCAHeader.String()
}
