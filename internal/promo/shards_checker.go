package promo

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
)

const (
	shardEntrySize = 11
	shardMaxCodeLen = 10
	shardMinCodeLen = 8
)

// ShardsChecker validates promo codes by checking membership in a single
// precomputed FNV-sharded, sorted shard file.
//
// It matches the output format of cmd/preprocessor_seq:
//   record: 1 byte len (8..10) + 10 bytes left-aligned code (zero padded)
//   records are sorted lexicographically by the code string.
type ShardsChecker struct {
	dir string

	mu sync.Mutex
	// Lazily loaded mmap slices per shard.
	shardData [ShardNumShards][]byte
	shardUnmap [ShardNumShards]func() error
	shardLoaded [ShardNumShards]bool
}

// LoadShardsPromo creates a ShardsChecker for shards in dir.
// Shards are mmap-loaded lazily on first use per shard.
func LoadShardsPromo(dir string) (*ShardsChecker, error) {
	if dir == "" {
		return nil, fmt.Errorf("shards dir is empty")
	}
	return &ShardsChecker{dir: dir}, nil
}

func (c *ShardsChecker) ensureShardLoaded(shard int) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if shard < 0 || shard >= ShardNumShards {
		return fmt.Errorf("shard out of range: %d", shard)
	}
	if c.shardLoaded[shard] {
		return nil
	}
	path := filepath.Join(c.dir, fmt.Sprintf("%03d.bin", shard))
	raw, unmap, err := mmapRead(path)
	if err != nil {
		// Missing shard means "not found".
		if os.IsNotExist(err) {
			c.shardLoaded[shard] = true
			c.shardData[shard] = nil
			c.shardUnmap[shard] = nil
			return nil
		}
		return err
	}

	c.shardLoaded[shard] = true
	c.shardData[shard] = raw
	c.shardUnmap[shard] = unmap
	return nil
}

func searchShardForKey(shardBytes []byte, key []byte) bool {
	if shardBytes == nil {
		return false
	}
	if len(key) < shardMinCodeLen || len(key) > shardMaxCodeLen {
		return false
	}
	if len(shardBytes)%shardEntrySize != 0 {
		// Corrupt shard.
		return false
	}
	n := len(shardBytes) / shardEntrySize

	lo, hi := 0, n
	for lo < hi {
		mid := (lo + hi) / 2
		base := mid * shardEntrySize
		l := int(shardBytes[base])
		if l < shardMinCodeLen || l > shardMaxCodeLen {
			return false
		}
		entryCode := shardBytes[base+1 : base+1+l]
		cmp := bytes.Compare(entryCode, key)
		if cmp == 0 {
			return true
		}
		if cmp < 0 {
			lo = mid + 1
		} else {
			hi = mid
		}
	}
	return false
}

func (c *ShardsChecker) Valid(code string) bool {
	if c == nil {
		return false
	}
	if !CouponCodePreludeOK(code) {
		return false
	}
	w := len(code)
	if w < shardMinCodeLen || w > shardMaxCodeLen {
		return false
	}
	key := []byte(code)
	sh := ShardIndexFNV256Bytes(key)
	if err := c.ensureShardLoaded(sh); err != nil {
		log.Printf("promo shards: failed to load shard %d: %v", sh, err)
		return false
	}

	c.mu.Lock()
	shardBytes := c.shardData[sh]
	c.mu.Unlock()
	return searchShardForKey(shardBytes, key)
}

func (c *ShardsChecker) Close() error {
	if c == nil {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	for i := 0; i < ShardNumShards; i++ {
		if c.shardUnmap[i] != nil {
			_ = c.shardUnmap[i]()
			c.shardUnmap[i] = nil
		}
		c.shardData[i] = nil
		c.shardLoaded[i] = false
	}
	return nil
}

