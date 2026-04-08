package promo

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteAndValidateManifest(t *testing.T) {
	dir := t.TempDir()
	if err := BuildShardsFromGzipDir(couponDirManifest(t), dir); err != nil {
		t.Fatal(err)
	}
	m, err := ReadManifest(dir)
	if err != nil || m == nil {
		t.Fatalf("manifest: %v", m)
	}
	if err := ValidateManifest(dir, m); err != nil {
		t.Fatal(err)
	}
}

func TestValidateManifest_SizeMismatch(t *testing.T) {
	dir := t.TempDir()
	if err := BuildShardsFromGzipDir(couponDirManifest(t), dir); err != nil {
		t.Fatal(err)
	}
	m, err := ReadManifest(dir)
	if err != nil || m == nil {
		t.Fatal(err)
	}
	m.Shards[0].Bytes++
	if err := ValidateManifest(dir, m); err == nil {
		t.Fatal("expected error on size mismatch")
	}
}

func TestValidateManifest_CorruptShardSize(t *testing.T) {
	dir := t.TempDir()
	if err := BuildShardsFromGzipDir(couponDirManifest(t), dir); err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(dir, "001.bin")
	if err := os.WriteFile(p, []byte("short"), 0o644); err != nil {
		t.Fatal(err)
	}
	m, err := ReadManifest(dir)
	if err != nil || m == nil {
		t.Fatal(err)
	}
	if err := ValidateManifest(dir, m); err == nil {
		t.Fatal("expected error on corrupt shard")
	}
}

func couponDirManifest(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	writeGZ(t, filepath.Join(dir, "couponbase1.gz"), "HAPPYHRS\n")
	writeGZ(t, filepath.Join(dir, "couponbase2.gz"), "XHAPPYHRS\n")
	writeGZ(t, filepath.Join(dir, "couponbase3.gz"), "ABHAPPYHRS\n")
	return dir
}
