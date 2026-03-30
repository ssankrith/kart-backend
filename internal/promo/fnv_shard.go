package promo

// ShardNumShards is the number of FNV-sharded buckets (must match preprocessor output).
const ShardNumShards = 256

// ShardIndexFNV256 returns a stable shard index in [0,255) for code.
// Must stay in sync with the offline preprocessor and any server-side lookup.
func ShardIndexFNV256(code string) int {
	h := uint32(2166136261)
	for i := 0; i < len(code); i++ {
		h ^= uint32(code[i])
		h *= 16777619
	}
	return int(h % uint32(ShardNumShards))
}

// ShardIndexFNV256Bytes returns a stable shard index in [0,255) for code bytes.
// Must stay in sync with ShardIndexFNV256(string).
func ShardIndexFNV256Bytes(code []byte) int {
	h := uint32(2166136261)
	for i := 0; i < len(code); i++ {
		h ^= uint32(code[i])
		h *= 16777619
	}
	return int(h % uint32(ShardNumShards))
}
