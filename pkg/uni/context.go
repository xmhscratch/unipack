package uni

import (
	"archive/tar"
	"sync"

	"github.com/hanwen/go-fuse/v2/fs"
)

var copyBuffer []byte = make([]byte, 1024)

// TarFile is a file read from a tar archive.
type TarFile struct {
	fs.Inode

	MountPath string
	FilePath  string
	Header    *tar.Header

	mu   sync.Mutex
	data []byte
}

// We decompress the file on demand in Open
var _ = (fs.NodeOpener)((*TarFile)(nil))

// Getattr sets the minimum, which is the size. A more full-featured
// FS would also set timestamps and permissions.
var _ = (fs.NodeGetattrer)((*TarFile)(nil))

type VFSRoot struct {
	fs.Inode
	TarFile    string
	MainFile   string
	MountPoint string
}

// The root populates the tree in its OnAdd method
var _ = (fs.NodeOnAdder)((*VFSRoot)(nil))
