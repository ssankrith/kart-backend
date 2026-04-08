// Command genmanifest writes manifest.json for an existing shard directory (000.bin … 255.bin).
//
//	go run ./cmd/genmanifest -dir ./shards_seq
package main

import (
	"flag"
	"log"

	"github.com/ssankrith/kart-backend/internal/promo"
)

func main() {
	dir := flag.String("dir", "./shards_seq", "directory containing shard *.bin files")
	flag.Parse()
	if err := promo.WriteShardManifestFromDir(*dir); err != nil {
		log.Fatal(err)
	}
	log.Printf("wrote %s/%s", *dir, promo.ManifestFileName)
}
