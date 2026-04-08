// Command dump_shard_codes reads final preprocessor shard bins (000.bin … 255.bin)
// and writes all entries that match runtime promo validation to one text file.
//
// Usage:
//
//	go run ./cmd/dump_shard_codes -dir ./shards_seq -out all_codes.txt
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ssankrith/kart-backend/internal/promo"
)

const (
	entrySize = 11
	minLen    = 8
	maxLen    = 10
)

func main() {
	dir := flag.String("dir", "shards_seq", "directory containing 000.bin … 255.bin")
	outPath := flag.String("out", "", "output text file (one code per line); default: stdout")
	flag.Parse()

	out := os.Stdout
	if *outPath != "" {
		f, err := os.Create(*outPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "dump_shard_codes: create %s: %v\n", *outPath, err)
			os.Exit(1)
		}
		defer f.Close()
		out = f
	}

	w := bufio.NewWriterSize(out, 256*1024)
	defer w.Flush()

	var written int64
	for shard := 0; shard < promo.ShardNumShards; shard++ {
		path := filepath.Join(*dir, fmt.Sprintf("%03d.bin", shard))
		data, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			fmt.Fprintf(os.Stderr, "dump_shard_codes: read %s: %v\n", path, err)
			os.Exit(1)
		}
		if len(data)%entrySize != 0 {
			fmt.Fprintf(os.Stderr, "dump_shard_codes: %s: size %d not multiple of %d\n", path, len(data), entrySize)
			os.Exit(1)
		}
		for off := 0; off < len(data); off += entrySize {
			rec := data[off : off+entrySize]
			l := int(rec[0])
			if l < minLen || l > maxLen {
				continue
			}
			for i := 1 + l; i < entrySize; i++ {
				if rec[i] != 0 {
					l = -1
					break
				}
			}
			if l < 0 {
				continue
			}
			code := string(rec[1 : 1+l])
			if !promo.CouponCodePreludeOK(code) {
				continue
			}
			if promo.ShardIndexFNV256(code) != shard {
				// Corrupt or wrong shard; skip.
				continue
			}
			if _, err := w.WriteString(code); err != nil {
				fmt.Fprintf(os.Stderr, "dump_shard_codes: write: %v\n", err)
				os.Exit(1)
			}
			if err := w.WriteByte('\n'); err != nil {
				fmt.Fprintf(os.Stderr, "dump_shard_codes: write: %v\n", err)
				os.Exit(1)
			}
			written++
		}
	}

	if *outPath == "" {
		fmt.Fprintf(os.Stderr, "dump_shard_codes: wrote %d codes\n", written)
	} else {
		fmt.Fprintf(os.Stderr, "dump_shard_codes: wrote %d codes to %s\n", written, *outPath)
	}
}
