// Command preprocessor_seq builds 256 shard files of valid promo codes (≥2 corpora)
// using a sequential, disk-backed pipeline:
// 1) scan each gzip sequentially and write window strings to per-corpus/per-shard temp files
// 2) dedup each temp file with OS sort -u
// 3) merge the 3 sorted unique streams per shard and emit final 11-byte entries
//
// Sliding-window semantics match docs/PROMO_DESIGN.md and internal/promo (8/9/10-byte windows per line):
// for each non-empty line enumerate every contiguous byte window of length 8, 9, 10.
//
// Example:
//   go run ./cmd/preprocessor_seq -data ./data -out ./shards -tmp ./tmp
package main

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/ssankrith/kart-backend/internal/promo"
)

func main() {
	dataDir := flag.String("data", "", "directory containing couponbase1.gz, couponbase2.gz, couponbase3.gz")
	outDir := flag.String("out", "./shards_seq", "output directory for 000.bin … 255.bin")
	tmpDir := flag.String("tmp", "", "temp directory (default: <out>/tmp)")
	scanLogEveryLines := flag.Int64("scanLogEveryLines", 10_000_000, "log scan progress every N non-empty lines per corpus")
	sortMemMB := flag.Int("sortMemMB", 0, "optional: memory hint to OS sort (0 = default). Best-effort; platform dependent.")
	flag.Parse()

	if *dataDir == "" {
		flag.Usage()
		os.Exit(2)
	}
	if *tmpDir == "" {
		*tmpDir = *outDir + "/tmp"
	}

	start := time.Now()
	log.Printf("promo-seq: start (data=%q out=%q tmp=%q)", *dataDir, *outDir, *tmpDir)

	if err := buildSequentialExternalSort(*dataDir, *outDir, *tmpDir, *scanLogEveryLines, *sortMemMB); err != nil {
		log.Fatal(err)
	}
	if err := promo.WriteShardManifestFromDir(*outDir); err != nil {
		log.Fatalf("manifest: %v", err)
	}

	log.Printf("promo-seq: done in %s", time.Since(start))
}

