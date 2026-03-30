package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

// buildSequentialExternalSort orchestrates:
// scan corpus1..3 sequentially -> dedup temp records with OS sort -u -> 3-way merge per shard.
func buildSequentialExternalSort(dataDir, outDir, tmpDir string, scanLogEveryLines int64, sortMemMB int) error {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(tmpDir, 0o755); err != nil {
		return err
	}

	start := time.Now()
	for corpusID := 1; corpusID <= 3; corpusID++ {
		log.Printf("promo-seq: corpus %d scan started...", corpusID)
		corpusTmpDir := filepath.Join(tmpDir, fmt.Sprintf("corpus%d", corpusID))
		_ = os.RemoveAll(corpusTmpDir)

		lines, bytes, err := scanCorpusToShardTemp(dataDir, corpusID, tmpDir, scanLogEveryLines)
		if err != nil {
			return fmt.Errorf("corpus %d scan: %w", corpusID, err)
		}
		log.Printf("promo-seq: corpus %d scan done lines=%d bytes=%d elapsed=%s", corpusID, lines, bytes, time.Since(start))

		log.Printf("promo-seq: corpus %d dedup started...", corpusID)
		if err := dedupCorpusShards(corpusTmpDir, sortMemMB); err != nil {
			return fmt.Errorf("corpus %d dedup: %w", corpusID, err)
		}
		log.Printf("promo-seq: corpus %d dedup done", corpusID)
	}

	log.Printf("promo-seq: merge started (256 shards)...")
	var totalWritten int64
	mergeStart := time.Now()
	lastLog := mergeStart
	for shard := 0; shard < 256; shard++ {
		if err := mergeShardAcrossCorpora(tmpDir, outDir, shard, &totalWritten); err != nil {
			return fmt.Errorf("merge shard %d: %w", shard, err)
		}
		if shard%16 == 15 || time.Since(lastLog) >= 30*time.Second {
			lastLog = time.Now()
			log.Printf("promo-seq: merged %d/%d shards elapsed=%s written=%d",
				shard+1, 256, time.Since(mergeStart), totalWritten)
		}
	}
	log.Printf("promo-seq: merge done shards=%d elapsed=%s totalWritten=%d", 256, time.Since(mergeStart), totalWritten)
	return nil
}

