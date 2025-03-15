package nstools_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/mrmarble/nstools"
)

func TestOpenPFS0(t *testing.T) {
	packed, err := os.OpenFile("testdata/packed.pfs0", os.O_RDONLY, 0644)
	if err != nil {
		t.Fatal(err)
	}
	defer packed.Close()

	pfs0, err := nstools.OpenPFS0(packed)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("PFS0: %+v", pfs0)

	if len(pfs0.Files) != 2 {
		t.Fatalf("expected 2 files, got %d", len(pfs0.Files))
	}

	if pfs0.Files[0].Name != "some_file.txt" {
		t.Fatalf("expected some_file.txt, got %s", pfs0.Files[1].Name)
	}

	if pfs0.Files[1].Name != "other_file.md" {
		t.Fatalf("expected other_file.md, got %s", pfs0.Files[0].Name)
	}
}

func TestPack(t *testing.T) {
	output := bytes.NewBuffer(nil)
	err := nstools.Pack(output, []string{
		"testdata/some_file.txt",
		"testdata/other_file.md",
	})

	if err != nil {
		t.Fatal(err)
	}

	pfs0, err := nstools.OpenPFS0(bytes.NewReader(output.Bytes()))
	if err != nil {
		t.Fatal(err)
	}

	if len(pfs0.Files) != 2 {
		t.Fatalf("expected 2 files, got %d", len(pfs0.Files))
	}
}

func TestUnpack(t *testing.T) {
	f, err := os.OpenFile("testdata/packed.pfs0", os.O_RDONLY, 0644)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	tmpDir, err := os.MkdirTemp("testdata", "unpacked")
	defer os.RemoveAll(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	err = nstools.Unpack(f, tmpDir)
	if err != nil {
		t.Fatal(err)
	}
}
