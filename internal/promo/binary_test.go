package promo

import (
	"math/bits"
	"path/filepath"
	"testing"
)

func TestLoadPromoLoadsShardsOnly(t *testing.T) {
	dir := t.TempDir()
	writeGZ(t, filepath.Join(dir, "couponbase1.gz"), "HAPPYHRS\n")
	writeGZ(t, filepath.Join(dir, "couponbase2.gz"), "XHAPPYHRS\n")
	writeGZ(t, filepath.Join(dir, "couponbase3.gz"), "ABHAPPYHRS\n")

	shardsDir := filepath.Join(dir, "shards_seq")
	if err := BuildShardsFromGzipDir(dir, shardsDir); err != nil {
		t.Fatal(err)
	}

	t.Setenv("PROMO_SHARDS_DIR", shardsDir)
	pc, err := LoadPromo(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer pc.Close()

	// Sanity check: using bitmask semantics, HAPPYHRS should be present in all 3 corpora
	// in this crafted fixture (as a window inside each line).
	counts, err := BuildBitmaskMapFromGzipDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	v := counts["HAPPYHRS"]
	if bits.OnesCount8(v) < 2 {
		t.Fatalf("test data broken: expected HAPPYHRS in ≥2 corpora, bits=%b", v)
	}

	if !pc.Valid("HAPPYHRS") {
		t.Fatal("expected HAPPYHRS valid via shards loader")
	}
}
