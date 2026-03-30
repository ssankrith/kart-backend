package main

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"

	"github.com/ssankrith/kart-backend/internal/promo"
)

const scanHeartbeatInterval = 30 * time.Second

func trimCRBytes(b []byte) []byte {
	if len(b) > 0 && b[len(b)-1] == '\r' {
		return b[:len(b)-1]
	}
	return b
}

func scanCorpusToShardTemp(dataDir string, corpusID int, tmpDir string, scanLogEveryLines int64) (lines int64, bytes int64, err error) {
	corpusPath := filepath.Join(dataDir, fmt.Sprintf("couponbase%d.gz", corpusID))
	in, err := os.Open(corpusPath)
	if err != nil {
		return 0, 0, err
	}
	defer in.Close()

	gz, err := gzip.NewReader(in)
	if err != nil {
		return 0, 0, err
	}
	defer gz.Close()

	// Create one writer per shard. This is simple and fast; if you hit FD limits, cap writers.
	corpusTmpDir := filepath.Join(tmpDir, fmt.Sprintf("corpus%d", corpusID))
	if err := os.MkdirAll(corpusTmpDir, 0o755); err != nil {
		return 0, 0, err
	}

	type shardWriter struct {
		f *os.File
		w *bufio.Writer
	}
	writers := make([]shardWriter, promo.ShardNumShards)
	for s := 0; s < promo.ShardNumShards; s++ {
		path := filepath.Join(corpusTmpDir, fmt.Sprintf("shard%03d.txt", s))
		f, err := os.Create(path)
		if err != nil {
			return 0, 0, err
		}
		writers[s] = shardWriter{
			f: f,
			w: bufio.NewWriterSize(f, 256*1024),
		}
	}
	defer func() {
		for s := 0; s < promo.ShardNumShards; s++ {
			if writers[s].w != nil {
				_ = writers[s].w.Flush()
			}
			if writers[s].f != nil {
				_ = writers[s].f.Close()
			}
		}
	}()

	sc := bufio.NewScanner(gz)
	sc.Buffer(make([]byte, 0, 64*1024), 64*1024)

	start := time.Now()
	lines, bytes = 0, 0
	nextLog := scanLogEveryLines

	var linesCount atomic.Int64
	var bytesCount atomic.Int64
	done := make(chan struct{})
	go func() {
		ticker := time.NewTicker(scanHeartbeatInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				l := linesCount.Load()
				if l == 0 {
					continue
				}
				b := bytesCount.Load()
				log.Printf("promo-seq: corpus %d scan heartbeat lines=%d bytes=%d elapsed=%s",
					corpusID, l, b, time.Since(start).Truncate(time.Second))
			case <-done:
				return
			}
		}
	}()

	for sc.Scan() {
		line := trimCRBytes(sc.Bytes())
		if len(line) == 0 {
			continue
		}
		lines++
		bytes += int64(len(line))
		linesCount.Store(lines)
		bytesCount.Store(bytes)

		// Sliding windows 8/9/10 over the *line* bytes.
		for win := 8; win <= 10; win++ {
			if len(line) < win {
				continue
			}
			for i := 0; i+win <= len(line); i++ {
				code := line[i : i+win]
				s := promo.ShardIndexFNV256Bytes(code)
				writers[s].w.Write(code)
				writers[s].w.WriteByte('\n')
			}
		}

		if scanLogEveryLines > 0 && lines == nextLog {
			el := time.Since(start)
			log.Printf("promo-seq: corpus %d scan progress lines=%d bytes=%d elapsed=%s", corpusID, lines, bytes, el)
			nextLog += scanLogEveryLines
		}
	}
	if err := sc.Err(); err != nil && err != io.EOF {
		close(done)
		return lines, bytes, err
	}

	close(done)
	// Final flush.
	for s := 0; s < promo.ShardNumShards; s++ {
		if err := writers[s].w.Flush(); err != nil {
			return lines, bytes, err
		}
		if err := writers[s].f.Close(); err != nil {
			return lines, bytes, err
		}
		writers[s].w = nil
		writers[s].f = nil
	}

	return lines, bytes, nil
}

