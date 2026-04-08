package promo

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadPromo_StrictRequiresManifest(t *testing.T) {
	dir := t.TempDir()
	writeGZBench(t, filepath.Join(dir, "couponbase1.gz"), "HAPPYHRS\n")
	writeGZBench(t, filepath.Join(dir, "couponbase2.gz"), "XHAPPYHRS\n")
	writeGZBench(t, filepath.Join(dir, "couponbase3.gz"), "ABHAPPYHRS\n")
	shardsDir := filepath.Join(dir, "shards_seq")
	if err := BuildShardsFromGzipDir(dir, shardsDir); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(shardsDir, ManifestFileName)); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PROMO_SHARDS_DIR", shardsDir)
	t.Setenv("PROMO_SHARDS_STRICT", "1")
	if _, err := LoadPromo(dir); err == nil {
		t.Fatal("expected error when manifest missing in strict mode")
	}
}
