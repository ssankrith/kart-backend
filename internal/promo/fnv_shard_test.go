package promo

import (
	"testing"
)

func TestShardIndexFNV256_deterministic(t *testing.T) {
	s := "HAPPYHRS"
	a := ShardIndexFNV256(s)
	b := ShardIndexFNV256(s)
	if a != b {
		t.Fatalf("expected stable shard, got %d vs %d", a, b)
	}
	if a < 0 || a >= ShardNumShards {
		t.Fatalf("shard out of range: %d", a)
	}
}

func TestShardIndexFNV256_range(t *testing.T) {
	for _, s := range []string{"", "A", "ABCDEFGH", "ABCDEFGHI", "ABCDEFGHIJ"} {
		i := ShardIndexFNV256(s)
		if i < 0 || i >= ShardNumShards {
			t.Fatalf("%q -> %d", s, i)
		}
	}
}

func TestShardIndexFNV256Bytes_matchesString(t *testing.T) {
	for _, s := range []string{"", "A", "HAPPYHRS", "ABCDEFGHI", "ABCDEFGHIJ"} {
		gotStr := ShardIndexFNV256(s)
		gotBytes := ShardIndexFNV256Bytes([]byte(s))
		if gotStr != gotBytes {
			t.Fatalf("mismatch for %q: string=%d bytes=%d", s, gotStr, gotBytes)
		}
	}
}
