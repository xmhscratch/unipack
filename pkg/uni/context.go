package uni

import (
	"archive/tar"
	"sync"

	"github.com/hanwen/go-fuse/v2/fs"
)

type FileInode struct {
	Header    *tar.Header
	MountPath string
	FilePath  string
	Content   []byte
}

// TarFile is a file read from a tar archive.
type TarFile struct {
	fs.Inode
	*FileInode

	vfs *VFSRoot
	mu  sync.Mutex
}

// We decompress the file on demand in Open
var _ = (fs.NodeOpener)((*TarFile)(nil))

// Getattr sets the minimum, which is the size. A more full-featured
// FS would also set timestamps and permissions.
var _ = (fs.NodeGetattrer)((*TarFile)(nil))

// =====================================================
type TFilesMap (map[uint64]*FileInode)

// FilesMap is a populated files content from a tar archive.
type FilesMap struct {
	uint64
	*TFilesMap
}

type FilesMapOpener interface {
	Load() (isNew bool, err error)
	Save() error
}

var _ = (FilesMapOpener)((*FilesMap)(nil))

// =====================================================
type VFSRoot struct {
	fs.Inode
	VFSRootOpener

	TarFile    string
	MainFile   string
	MountPoint string
	HashSum    string
	FilesMap   *FilesMap
	IsNew      bool

	readyMount chan struct{}
}

type VFSRootOpener interface {
	Init() (err error)
	Walk(func(h *tar.Header, r *tar.Reader) error) error
}

// The root populates the tree in its OnAdd method
var _ = (fs.NodeOnAdder)((*VFSRoot)(nil))

var _ = (VFSRootOpener)((*VFSRoot)(nil))
