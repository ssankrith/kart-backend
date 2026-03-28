package promo

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"unicode/utf8"
)

// Checker validates coupon codes against gzipped corpora (Oolio rules).
type Checker struct {
	blobs [][]byte
}

// LoadFromGZIPFiles decompresses each file into memory once at startup.
func LoadFromGZIPFiles(paths []string) (*Checker, error) {
	if len(paths) != 3 {
		return nil, fmt.Errorf("expected 3 coupon files, got %d", len(paths))
	}
	blobs := make([][]byte, 0, 3)
	for _, p := range paths {
		b, err := readGZIPFile(p)
		if err != nil {
			return nil, fmt.Errorf("coupon file %q: %w", p, err)
		}
		blobs = append(blobs, b)
	}
	return &Checker{blobs: blobs}, nil
}

func readGZIPFile(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return nil, err
	}
	defer gz.Close()
	return io.ReadAll(gz)
}

// Valid reports whether code satisfies:
// 1) UTF-8 length between 8 and 10 inclusive
// 2) substring appears in at least two of the three corpora.
func (c *Checker) Valid(code string) bool {
	n := utf8.RuneCountInString(code)
	if n < 8 || n > 10 {
		return false
	}
	if len(c.blobs) < 3 {
		return false
	}
	needle := []byte(code)
	matches := 0
	for _, blob := range c.blobs {
		if bytes.Contains(blob, needle) {
			matches++
		}
	}
	return matches >= 2
}

// DirPaths returns default filenames under dir (couponbase1.gz .. 3).
func DirPaths(dir string) []string {
	return []string{
		filepath.Join(dir, "couponbase1.gz"),
		filepath.Join(dir, "couponbase2.gz"),
		filepath.Join(dir, "couponbase3.gz"),
	}
}
