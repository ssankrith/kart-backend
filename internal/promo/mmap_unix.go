//go:build !windows

package promo

import (
	"os"

	"golang.org/x/sys/unix"
)

func mmapRead(path string) ([]byte, func() error, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	st, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, nil, err
	}
	sz := int(st.Size())
	if sz == 0 {
		f.Close()
		return nil, func() error { return nil }, nil
	}
	data, err := unix.Mmap(int(f.Fd()), 0, sz, unix.PROT_READ, unix.MAP_PRIVATE)
	f.Close()
	if err != nil {
		return nil, nil, err
	}
	return data, func() error { return unix.Munmap(data) }, nil
}
