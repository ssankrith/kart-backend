package promo

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"os"
	"path/filepath"
)

// BuildBitmaskMapFromGzipDir scans couponbase{1,2,3}.gz under dataDir with the same
// sliding-window rules as LoadMapFromDataDir (widths 8, 9, 10 per line) and returns the
// merged map[string]uint8 bitmask. Intended for offline tools (e.g. shard preprocessor).
func BuildBitmaskMapFromGzipDir(dataDir string) (map[string]uint8, error) {
	paths := []string{
		filepath.Join(dataDir, "couponbase1.gz"),
		filepath.Join(dataDir, "couponbase2.gz"),
		filepath.Join(dataDir, "couponbase3.gz"),
	}

	out := make(map[string]uint8, 1024)
	for i, p := range paths {
		corpusID := i + 1
		f, err := os.Open(p)
		if err != nil {
			return nil, fmt.Errorf("couponbase%d: %w", corpusID, err)
		}
		gz, err := gzip.NewReader(f)
		if err != nil {
			_ = f.Close()
			return nil, fmt.Errorf("couponbase%d: %w", corpusID, err)
		}

		// Local dedup for this corpus only.
		local := make(map[string]struct{}, 1024)
		sc := bufio.NewScanner(gz)
		sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
		for sc.Scan() {
			line := trimCRBytes(sc.Bytes())
			if len(line) == 0 {
				continue
			}
			if len(line) >= 8 {
				for j := 0; j+8 <= len(line); j++ {
					local[string(line[j:j+8])] = struct{}{}
				}
			}
			if len(line) >= 9 {
				for j := 0; j+9 <= len(line); j++ {
					local[string(line[j:j+9])] = struct{}{}
				}
			}
			if len(line) >= 10 {
				for j := 0; j+10 <= len(line); j++ {
					local[string(line[j:j+10])] = struct{}{}
				}
			}
		}
		scErr := sc.Err()
		_ = gz.Close()
		_ = f.Close()
		if scErr != nil {
			return nil, fmt.Errorf("couponbase%d: %w", corpusID, scErr)
		}

		bit := uint8(1 << uint(i))
		for code := range local {
			out[code] |= bit
		}
	}
	return out, nil
}

func trimCRBytes(b []byte) []byte {
	if len(b) > 0 && b[len(b)-1] == '\r' {
		return b[:len(b)-1]
	}
	return b
}
