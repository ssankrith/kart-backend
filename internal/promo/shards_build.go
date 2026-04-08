package promo

import (
	"bufio"
	"fmt"
	"math/bits"
	"os"
	"path/filepath"
	"sort"
)

const shardBuildEntrySize = 11

// BuildShardsFromGzipDir builds FNV-sharded promo shard files in outDir.
//
// It is intended for tests and small offline generation, not for huge corpora:
// it uses an in-memory map for deduplication.
//
// Output format must match cmd/preprocessor_seq and ShardsChecker:
// each record is 11 bytes: [len(8..10)][code bytes left-aligned, zero padded].
func BuildShardsFromGzipDir(dataDir, outDir string) error {
	counts, err := BuildBitmaskMapFromGzipDir(dataDir)
	if err != nil {
		return err
	}

	buckets := make([][]string, ShardNumShards)
	for code, v := range counts {
		if bits.OnesCount8(v) < 2 {
			continue
		}
		s := ShardIndexFNV256(code)
		buckets[s] = append(buckets[s], code)
	}

	for i := range buckets {
		sort.Strings(buckets[i])
	}

	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}

	for shard := 0; shard < ShardNumShards; shard++ {
		path := filepath.Join(outDir, fmt.Sprintf("%03d.bin", shard))
		f, err := os.Create(path)
		if err != nil {
			return err
		}
		w := bufio.NewWriterSize(f, 256*1024)
		var entry [shardBuildEntrySize]byte
		for _, code := range buckets[shard] {
			entry = [shardBuildEntrySize]byte{}
			entry[0] = byte(len(code))
			copy(entry[1:], code)
			if _, err := w.Write(entry[:]); err != nil {
				_ = f.Close()
				return err
			}
		}
		if err := w.Flush(); err != nil {
			_ = f.Close()
			return err
		}
		if err := f.Close(); err != nil {
			return err
		}
	}

	return WriteShardManifestFromDir(outDir)
}

