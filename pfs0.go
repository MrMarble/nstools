package nstools

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Partition is a file header in a PFS0 container.
// Use the Extract method to extract the file data.
type Partition struct {
	Name   string
	Size   int64
	Offset int64
}

// PFS0 is a file system, a container that holds multiple files.
type PFS0 struct {
	data       io.ReadSeeker
	Magic      [4]byte
	HeaderSize uint32
	Files      []Partition
}

// OpenPFS0 opens a PFS0 container from the provided reader.
// Only the header is read, use the Extract method to extract files.
func OpenPFS0(data io.ReadSeeker) (*PFS0, error) {
	var magic [4]byte
	binary.Read(data, binary.LittleEndian, &magic)

	if magic != MagicPFS0 {
		return nil, fmt.Errorf("%w: got %s, expected PFS0", ErrInvalidMagic, string(magic[:]))
	}

	pfs0 := &PFS0{Magic: magic, data: data}
	pfs0.open()
	return pfs0, nil
}

func (pfs0 *PFS0) Close() error {
	if closer, ok := pfs0.data.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

func (pfs0 *PFS0) open() {
	// Read header
	var fileCount, stringTableSize uint32

	binary.Read(pfs0.data, binary.LittleEndian, &fileCount)
	binary.Read(pfs0.data, binary.LittleEndian, &stringTableSize)
	pfs0.data.Seek(4, io.SeekCurrent) // Junk data

	stringTableOffset := 0x10 + (int64(fileCount) * 0x18)
	pfs0.data.Seek(stringTableOffset, io.SeekStart)

	stringTableBytes := make([]byte, stringTableSize)
	pfs0.data.Read(stringTableBytes)
	stringTable := string(stringTableBytes)

	stringEndOffset := stringTableSize
	pfs0.HeaderSize = 0x10 + (fileCount * 0x18) + stringTableSize

	// Read files
	pfs0.Files = make([]Partition, 0, fileCount)
	for i := range int(fileCount) {
		j := fileCount - uint32(i) - 1
		pfs0.data.Seek(int64(0x10+j*0x18), io.SeekStart)

		var offset, size int64
		var nameOffset uint32

		binary.Read(pfs0.data, binary.LittleEndian, &offset)
		binary.Read(pfs0.data, binary.LittleEndian, &size)
		binary.Read(pfs0.data, binary.LittleEndian, &nameOffset)

		name := strings.Trim(stringTable[nameOffset:stringEndOffset], "\x00")
		stringEndOffset = nameOffset

		pfs0.Files = append(pfs0.Files, Partition{name, size, offset + int64(pfs0.HeaderSize)})
	}

	// Sort files by offset
	sort.Slice(pfs0.Files, func(i, j int) bool {
		return pfs0.Files[i].Offset < pfs0.Files[j].Offset
	})
}

// Extract writes the file at the given index to the provided writer.
func (pfs0 *PFS0) Extract(w io.Writer, fileIndex int) error {
	if fileIndex < 0 || fileIndex >= len(pfs0.Files) {
		return fmt.Errorf("invalid file index %d with max %d", fileIndex, len(pfs0.Files))
	}
	file := pfs0.Files[fileIndex]

	pfs0.data.Seek(file.Offset, io.SeekStart)
	_, err := io.CopyN(w, pfs0.data, file.Size)
	return err
}

// Unpack extracts all files from the PFS0 stream to the provided output directory.
func Unpack(data io.ReadSeeker, output string) error {
	pfs0, err := OpenPFS0(data)
	if err != nil {
		return err
	}

	if len(pfs0.Files) == 0 {
		return fmt.Errorf("no files found in PFS0 container")
	}

	err = os.MkdirAll(output, 0755)
	if err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	for i, file := range pfs0.Files {
		filePath := fmt.Sprintf("%s/%s", output, file.Name)
		f, err := os.Create(filePath)
		if err != nil {
			return fmt.Errorf("failed to create file %s: %w", filePath, err)
		}

		err = pfs0.Extract(f, i)
		if err != nil {
			f.Close()
			return fmt.Errorf("failed to extract file %s: %w", filePath, err)
		}

		err = f.Close()
		if err != nil {
			return fmt.Errorf("failed to close file %s: %w", filePath, err)
		}
	}

	return nil
}

// Pack creates a PFS0 container from the provided files and writes it to the provided writer.
func Pack(w io.Writer, files []string) error {
	if len(files) == 0 {
		return fmt.Errorf("no files provided")
	}

	// Remove path from files
	baseFiles := make([]string, 0, len(files))
	for _, file := range files {
		baseFiles = append(baseFiles, filepath.Base(file))
	}

	// Calculate string table size
	stringTableSize := 0
	for _, file := range baseFiles {
		stringTableSize += len(file) + 1
	}

	// Write header
	binary.Write(w, binary.LittleEndian, MagicPFS0)
	binary.Write(w, binary.LittleEndian, uint32(len(baseFiles)))
	binary.Write(w, binary.LittleEndian, uint32(stringTableSize))
	binary.Write(w, binary.LittleEndian, uint32(0)) // Padding

	// Write files
	stringTable := make([]byte, 0)
	var offset int64 = 0

	for i, file := range files {
		f, err := os.Open(file)
		if err != nil {
			return fmt.Errorf("failed to open file %s: %w", file, err)
		}

		stat, err := f.Stat()
		if err != nil {
			return fmt.Errorf("failed to stat file %s: %w", file, err)
		}

		binary.Write(w, binary.LittleEndian, offset)
		binary.Write(w, binary.LittleEndian, stat.Size())
		binary.Write(w, binary.LittleEndian, uint32(len(stringTable)))
		binary.Write(w, binary.LittleEndian, uint32(0)) // Padding

		stringTable = append(stringTable, []byte(baseFiles[i])...)
		stringTable = append(stringTable, 0)

		offset += stat.Size()
	}

	// Write string table
	binary.Write(w, binary.LittleEndian, stringTable)

	// Write files
	for _, file := range files {
		f, err := os.Open(file)
		if err != nil {
			return fmt.Errorf("failed to open file %s: %w", file, err)
		}

		_, err = io.Copy(w, f)
		if err != nil {
			return fmt.Errorf("failed to write file %s: %w", file, err)
		}
	}
	binary.Write(w, binary.LittleEndian, uint8(0)) // Padding

	return nil
}
