package main

import (
	"compress/gzip"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/ssankrith/kart-backend/internal/promo"
)

func writeGZ(t *testing.T, path string, content string) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	gz := gzip.NewWriter(f)
	if _, err := gz.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
}

func decodeShardEntries(t *testing.T, shardPath string) []string {
	t.Helper()
	data, err := os.ReadFile(shardPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(data)%11 != 0 {
		t.Fatalf("shard file size %d not divisible by 11: %s", len(data), shardPath)
	}
	out := make([]string, 0, len(data)/11)
	for off := 0; off < len(data); off += 11 {
		l := int(data[off])
		if l < 8 || l > 10 {
			t.Fatalf("unexpected entry length byte=%d in %s", l, shardPath)
		}
		codeBytes := data[off+1 : off+1+10]
		code := string(codeBytes[:l])
		out = append(out, code)
	}
	return out
}

func TestPreprocessorSeq_EndToEnd_Tiny(t *testing.T) {
	dataDir := t.TempDir()
	outDir := filepath.Join(t.TempDir(), "out")
	tmpDir := filepath.Join(t.TempDir(), "tmp")

	// Construct 3 lines such that "HAPPYHRS" (len=8) appears as a substring in all three corpora.
	// corpus1: exact 8-char line
	writeGZ(t, filepath.Join(dataDir, "couponbase1.gz"), "HAPPYHRS\n")
	// corpus2: 9 chars, contains HAPPYHRS as an 8-char window starting at index 1
	writeGZ(t, filepath.Join(dataDir, "couponbase2.gz"), "XHAPPYHRS\n")
	// corpus3: 10 chars, contains HAPPYHRS as an 8-char window starting at index 2
	writeGZ(t, filepath.Join(dataDir, "couponbase3.gz"), "ABHAPPYHRS\n")

	if err := buildSequentialExternalSort(dataDir, outDir, tmpDir, 0, 0); err != nil {
		t.Fatal(err)
	}

	code := "HAPPYHRS"
	sh := promo.ShardIndexFNV256(code)
	shardPath := filepath.Join(outDir, fmt.Sprintf("%03d.bin", sh))

	codes := decodeShardEntries(t, shardPath)
	if len(codes) == 0 {
		t.Fatalf("expected at least one valid code in shard %d", sh)
	}

	// Must contain HAPPYHRS.
	found := false
	for _, c := range codes {
		if c == code {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected %q in %s entries=%v", code, shardPath, codes)
	}

	// Ensure codes are sorted lexicographically.
	sorted := append([]string(nil), codes...)
	sort.Strings(sorted)
	if strings.Join(sorted, "\n") != strings.Join(codes, "\n") {
		t.Fatalf("entries not sorted in %s codes=%v sorted=%v", shardPath, codes, sorted)
	}

	// Ensure a code that only appears in one corpus doesn't show up.
	// Example: "XHAPPYHR" appears in corpus2 as the other 8-char window, but not in corpus1/corpus3.
	absent := "XHAPPYHR"
	for _, c := range codes {
		if c == absent {
			t.Fatalf("did not expect %q in %s", absent, shardPath)
		}
	}
}

