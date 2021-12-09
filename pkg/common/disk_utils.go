package common

import (
	"log"
	"syscall"
)

type DiskInfo struct {
	Total uint64 `json:"total"`
	Free  uint64 `json:"free"`
}

func (d *DiskInfo) Used() uint64 {
	return d.Total - d.Free
}

const (
	B  = 1
	KB = 1024 * B
	MB = 1024 * KB
	GB = 1024 * MB
)

func DiskUsage(path string) (*DiskInfo, error) {

	fileState := syscall.Statfs_t{}
	if err := syscall.Statfs(path, &fileState); err != nil {
		log.Printf("disk usage error : %v\n", err)
		return nil, err
	}

	disInfo := &DiskInfo{
		Total: fileState.Blocks * uint64(fileState.Bsize),
		Free:  fileState.Bfree * uint64(fileState.Bsize),
	}

	return disInfo, nil
}