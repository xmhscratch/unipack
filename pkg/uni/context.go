package uni

import (
	"archive/tar"
	"sync"

	"github.com/hanwen/go-fuse/v2/fs"
)

// TarFile is a file read from a tar archive.
type TarFile struct {
	fs.Inode
	// reader *tar.Reader

	FilePath string
	Header   *tar.Header

	mu   sync.Mutex
	data []byte
}

// We decompress the file on demand in Open
var _ = (fs.NodeOpener)((*TarFile)(nil))

// Getattr sets the minimum, which is the size. A more full-featured
// FS would also set timestamps and permissions.
var _ = (fs.NodeGetattrer)((*TarFile)(nil))

type TarRoot struct {
	fs.Inode
	TarPath string
}

// The root populates the tree in its OnAdd method
var _ = (fs.NodeOnAdder)((*TarRoot)(nil))
