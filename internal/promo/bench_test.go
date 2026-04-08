package promo

import (
	"compress/gzip"
	"math/bits"
	"os"
	"path/filepath"
	"testing"
)

func benchCouponDir(b *testing.B) string {
	b.Helper()
	dir := b.TempDir()
	writeGZBench(b, filepath.Join(dir, "couponbase1.gz"), "HAPPYHRS\n")
	writeGZBench(b, filepath.Join(dir, "couponbase2.gz"), "XHAPPYHRS\n")
	writeGZBench(b, filepath.Join(dir, "couponbase3.gz"), "ABHAPPYHRS\n")
	return dir
}

func writeGZBench(tb testing.TB, path, content string) {
	tb.Helper()
	f, err := os.Create(path)
	if err != nil {
		tb.Fatal(err)
	}
	defer f.Close()
	gz := gzip.NewWriter(f)
	if _, err := gz.Write([]byte(content)); err != nil {
		tb.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		tb.Fatal(err)
	}
}

func mustBenchShards(b *testing.B) *ShardsChecker {
	b.Helper()
	dir := benchCouponDir(b)
	shardsDir := filepath.Join(dir, "shards_seq")
	if err := BuildShardsFromGzipDir(dir, shardsDir); err != nil {
		b.Fatal(err)
	}
	c, err := LoadShardsPromo(shardsDir)
	if err != nil {
		b.Fatal(err)
	}
	b.Cleanup(func() { _ = c.Close() })
	return c
}

// BenchmarkValid_WarmHit touches the shard once then measures repeated lookups.
func BenchmarkValid_WarmHit(b *testing.B) {
	c := mustBenchShards(b)
	_ = c.Valid("HAPPYHRS")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = c.Valid("HAPPYHRS")
	}
}

// BenchmarkValid_ColdFirstPay mmap + first lookup per iteration (new checker each time).
func BenchmarkValid_ColdFirstLookup(b *testing.B) {
	dir := benchCouponDir(b)
	shardsDir := filepath.Join(dir, "shards_seq")
	if err := BuildShardsFromGzipDir(dir, shardsDir); err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c, err := LoadShardsPromo(shardsDir)
		if err != nil {
			b.Fatal(err)
		}
		_ = c.Valid("HAPPYHRS")
		_ = c.Close()
	}
}

// BenchmarkValid_PreludeReject invalid length / UTF-8 — no shard access.
func BenchmarkValid_PreludeReject(b *testing.B) {
	c := mustBenchShards(b)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = c.Valid("SHORT")
	}
}

// BenchmarkValid_WarmMiss valid shape but not present in shards.
func BenchmarkValid_WarmMiss(b *testing.B) {
	c := mustBenchShards(b)
	code := "HAPPYHRX" // 8 chars, same length as corpus code but not in set
	if c.Valid(code) {
		b.Fatal("expected miss for benchmark code")
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = c.Valid(code)
	}
}

// BenchmarkValid_NaiveMap compares in-memory map membership (test-only baseline).
func BenchmarkValid_NaiveMap(b *testing.B) {
	dir := benchCouponDir(b)
	counts, err := BuildBitmaskMapFromGzipDir(dir)
	if err != nil {
		b.Fatal(err)
	}
	valid := make(map[string]struct{})
	for s, mask := range counts {
		if bits.OnesCount8(mask) >= 2 {
			valid[s] = struct{}{}
		}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = valid["HAPPYHRS"]
	}
}
