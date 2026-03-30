package promo

import (
	"math/bits"
	"path/filepath"
	"testing"
)

func TestBuildBitmaskMapFromGzipDir(t *testing.T) {
	dir := t.TempDir()
	writeGZ(t, filepath.Join(dir, "couponbase1.gz"), "HAPPYHRS\n")
	writeGZ(t, filepath.Join(dir, "couponbase2.gz"), "XHAPPYHRS\n")
	writeGZ(t, filepath.Join(dir, "couponbase3.gz"), "ABHAPPYHRS\n")

	m, err := BuildBitmaskMapFromGzipDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	v, ok := m["HAPPYHRS"]
	if !ok {
		t.Fatal("expected HAPPYHRS in map")
	}
	if bits.OnesCount8(v) < 2 {
		t.Fatalf("expected ≥2 corpora for HAPPYHRS, got bits=%b", v)
	}
}
