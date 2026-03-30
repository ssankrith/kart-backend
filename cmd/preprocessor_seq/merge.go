package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/ssankrith/kart-backend/internal/promo"
)

func dedupCorpusShards(corpusTmpDir string, sortMemMB int) error {
	// corpusTmpDir contains shard%03d.txt for that corpus.
	lastLog := time.Now()
	for shard := 0; shard < promo.ShardNumShards; shard++ {
		inPath := filepath.Join(corpusTmpDir, fmt.Sprintf("shard%03d.txt", shard))
		uniqPath := filepath.Join(corpusTmpDir, fmt.Sprintf("shard%03d.uniq", shard))

		st, err := os.Stat(inPath)
		if err != nil {
			return err
		}
		if st.Size() == 0 {
			// Ensure uniq file exists (may be empty).
			f, err := os.Create(uniqPath)
			if err != nil {
				return err
			}
			_ = f.Close()
			continue
		}

		// External dedup by lexicographic sort.
		// We rely on LC_ALL=C for bytewise ordering.
		args := []string{}
		if sortMemMB > 0 {
			// Best-effort memory hint; BSD/GNU sort accept different flags.
			// BSD sort supports `-S <size>`; GNU sort supports `--buffer-size=<size>`.
			// Keep this lightweight and portable by using BSD-style `-S` with MB suffix.
			args = append(args, "-S", fmt.Sprintf("%dM", sortMemMB))
		}
		args = append(args, "-u", "-o", uniqPath, inPath)
		cmd := exec.Command("sort", args...)
		cmd.Env = append(os.Environ(), "LC_ALL=C")
		start := time.Now()
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("sort -u failed for %s: %w (out=%s)", inPath, err, string(out))
		}
		el := time.Since(start)

		// Remove raw temp file to save disk.
		_ = os.Remove(inPath)

		if shard%16 == 15 || time.Since(lastLog) >= 30*time.Second {
			lastLog = time.Now()
			if fi, err := os.Stat(uniqPath); err == nil {
				log.Printf("promo-seq: dedup shard %d/256 corpus=%s elapsed=%s uniqSize=%d bytes",
					shard+1, filepath.Base(corpusTmpDir), el, fi.Size())
			}
		}
	}
	return nil
}

type lineStream struct {
	f   *os.File
	sc  *bufio.Scanner
	cur string
	ok  bool
}

func newLineStream(path string) (*lineStream, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 64*1024)
	ls := &lineStream{f: f, sc: sc}
	if sc.Scan() {
		ls.cur = sc.Text()
		ls.ok = true
	} else {
		ls.ok = false
	}
	return ls, nil
}

func (ls *lineStream) advance() {
	if ls.sc.Scan() {
		ls.cur = ls.sc.Text()
		ls.ok = true
	} else {
		ls.ok = false
	}
}

func (ls *lineStream) close() {
	_ = ls.f.Close()
}

func mergeShardAcrossCorpora(tmpDir, outDir string, shard int, totalWritten *int64) error {
	outPath := filepath.Join(outDir, fmt.Sprintf("%03d.bin", shard))
	outF, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer outF.Close()

	w := bufio.NewWriterSize(outF, 256*1024)
	defer w.Flush()

	streams := make([]*lineStream, 3)
	for ci := 0; ci < 3; ci++ {
		uniqPath := filepath.Join(tmpDir, fmt.Sprintf("corpus%d", ci+1), fmt.Sprintf("shard%03d.uniq", shard))
		ls, err := newLineStream(uniqPath)
		if err != nil {
			if os.IsNotExist(err) {
				// Treat missing shard as empty.
				streams[ci] = &lineStream{ok: false}
				continue
			}
			return err
		}
		streams[ci] = ls
	}
	defer func() {
		for _, s := range streams {
			if s != nil {
				s.close()
			}
		}
	}()

	for {
		minCode := ""
		minSet := false
		for ci := 0; ci < 3; ci++ {
			if streams[ci] != nil && streams[ci].ok {
				code := streams[ci].cur
				if !minSet || code < minCode {
					minCode = code
					minSet = true
				}
			}
		}
		if !minSet {
			break
		}

		// Count in how many corpora this exact code exists.
		count := 0
		for ci := 0; ci < 3; ci++ {
			if streams[ci] != nil && streams[ci].ok && streams[ci].cur == minCode {
				count++
				streams[ci].advance()
			}
		}

		if count >= 2 {
			var entry [11]byte
			if len(minCode) < 8 || len(minCode) > 10 {
				// Should not happen: we only emit windows of length 8/9/10 from scan.
				continue
			}
			entry[0] = byte(len(minCode))
			copy(entry[1:], minCode)
			if _, err := w.Write(entry[:]); err != nil {
				return err
			}
			*totalWritten++
		}
	}

	return nil
}

