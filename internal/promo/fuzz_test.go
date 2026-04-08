package promo

import (
	"path/filepath"
	"testing"
)

func FuzzValid(f *testing.F) {
	c := mustFuzzChecker(f)
	f.Add("HAPPYHRS")
	f.Add("SHORT")
	f.Add("ZZZZZZZZ")
	f.Fuzz(func(t *testing.T, code string) {
		_ = c.Valid(code)
	})
}

func mustFuzzChecker(f *testing.F) *ShardsChecker {
	f.Helper()
	dir := f.TempDir()
	writeGZBench(f, filepath.Join(dir, "couponbase1.gz"), "HAPPYHRS\n")
	writeGZBench(f, filepath.Join(dir, "couponbase2.gz"), "XHAPPYHRS\n")
	writeGZBench(f, filepath.Join(dir, "couponbase3.gz"), "ABHAPPYHRS\n")
	shardsDir := filepath.Join(dir, "s")
	if err := BuildShardsFromGzipDir(dir, shardsDir); err != nil {
		f.Fatal(err)
	}
	pc, err := LoadShardsPromo(shardsDir)
	if err != nil {
		f.Fatal(err)
	}
	f.Cleanup(func() { _ = pc.Close() })
	return pc
}
