package promo

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// ManifestFileName is written next to shard *.bin files.
const ManifestFileName = "manifest.json"

// ManifestFormatVersion bumps when on-disk layout or semantics change.
const ManifestFormatVersion = 1

// FNVVariant documents the shard routing hash (see fnv_shard.go).
const FNVVariant = "FNV-1a-32-mod-256"

// ShardManifest describes the preprocessor output for validation at startup.
type ShardManifest struct {
	FormatVersion int    `json:"format_version"`
	RecordSize    int    `json:"record_size"`
	ShardCount    int    `json:"shard_count"`
	FNVVariant    string `json:"fnv_variant"`
	BuiltAt       string `json:"built_at"`
	Shards        []struct {
		File   string `json:"file"`
		Bytes  int64  `json:"bytes"`
		SHA256 string `json:"sha256,omitempty"`
	} `json:"shards"`
}

// WriteShardManifestFromDir stats each 000.bin…255.bin under dir and writes manifest.json
// with SHA-256 of each file. Empty or missing shards are recorded as 0 bytes (file may be absent).
func WriteShardManifestFromDir(dir string) error {
	m := ShardManifest{
		FormatVersion: ManifestFormatVersion,
		RecordSize:    shardEntrySize,
		ShardCount:    ShardNumShards,
		FNVVariant:    FNVVariant,
		BuiltAt:       time.Now().UTC().Format(time.RFC3339),
	}

	for shard := 0; shard < ShardNumShards; shard++ {
		name := fmt.Sprintf("%03d.bin", shard)
		p := filepath.Join(dir, name)
		st, err := os.Stat(p)
		var sz int64
		var sum string
		if err == nil {
			sz = st.Size()
			if sz > 0 {
				h, err := fileSHA256(p)
				if err != nil {
					return fmt.Errorf("manifest: %s: %w", name, err)
				}
				sum = h
			}
		} else if os.IsNotExist(err) {
			sz = 0
		} else {
			return fmt.Errorf("manifest: stat %s: %w", name, err)
		}
		m.Shards = append(m.Shards, struct {
			File   string `json:"file"`
			Bytes  int64  `json:"bytes"`
			SHA256 string `json:"sha256,omitempty"`
		}{File: name, Bytes: sz, SHA256: sum})
	}

	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, ManifestFileName), data, 0o644)
}

func fileSHA256(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:]), nil
}

// ReadManifest loads manifest.json from dir, or returns nil if missing.
func ReadManifest(dir string) (*ShardManifest, error) {
	p := filepath.Join(dir, ManifestFileName)
	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var m ShardManifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

// ValidateManifest checks that shard files on disk match the manifest (sizes and optional hashes).
func ValidateManifest(dir string, m *ShardManifest) error {
	if m == nil {
		return fmt.Errorf("manifest: nil")
	}
	if m.FormatVersion != ManifestFormatVersion {
		return fmt.Errorf("manifest: unsupported format_version %d (want %d)", m.FormatVersion, ManifestFormatVersion)
	}
	if m.RecordSize != shardEntrySize {
		return fmt.Errorf("manifest: record_size %d (want %d)", m.RecordSize, shardEntrySize)
	}
	if m.ShardCount != ShardNumShards {
		return fmt.Errorf("manifest: shard_count %d (want %d)", m.ShardCount, ShardNumShards)
	}
	if m.FNVVariant != "" && m.FNVVariant != FNVVariant {
		return fmt.Errorf("manifest: unexpected fnv_variant %q (want %q)", m.FNVVariant, FNVVariant)
	}
	if len(m.Shards) != ShardNumShards {
		return fmt.Errorf("manifest: shards len %d (want %d)", len(m.Shards), ShardNumShards)
	}
	for i, ent := range m.Shards {
		p := filepath.Join(dir, ent.File)
		st, err := os.Stat(p)
		if err != nil {
			if os.IsNotExist(err) && ent.Bytes == 0 {
				continue
			}
			return fmt.Errorf("manifest: shard %d: %w", i, err)
		}
		if st.Size() != ent.Bytes {
			return fmt.Errorf("manifest: %s size %d (want %d per manifest)", ent.File, st.Size(), ent.Bytes)
		}
		if ent.Bytes > 0 && ent.Bytes%int64(shardEntrySize) != 0 {
			return fmt.Errorf("manifest: %s size %d not multiple of record size %d", ent.File, ent.Bytes, shardEntrySize)
		}
		if ent.SHA256 != "" {
			got, err := fileSHA256(p)
			if err != nil {
				return err
			}
			if got != ent.SHA256 {
				return fmt.Errorf("manifest: %s sha256 mismatch", ent.File)
			}
		}
	}
	return nil
}
