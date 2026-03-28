package promo

import (
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"
)

func TestChecker_Valid(t *testing.T) {
	dir := t.TempDir()
	writeGZ(t, filepath.Join(dir, "couponbase1.gz"), "aaa HAPPYHRS bbb")
	writeGZ(t, filepath.Join(dir, "couponbase2.gz"), "HAPPYHRS")
	writeGZ(t, filepath.Join(dir, "couponbase3.gz"), "no")

	c, err := LoadFromGZIPFiles(DirPaths(dir))
	if err != nil {
		t.Fatal(err)
	}
	if !c.Valid("HAPPYHRS") {
		t.Fatal("expected HAPPYHRS valid")
	}
	if c.Valid("SHORT") {
		t.Fatal("expected SHORT invalid (length)")
	}
	if c.Valid("SUPER10000") {
		t.Fatal("expected long code invalid")
	}
	if c.Valid("NOTTHERE") {
		t.Fatal("expected absent code invalid")
	}
}

func writeGZ(t *testing.T, path, content string) {
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
