package uni

import (
	"context"
	"path/filepath"
	"syscall"
	"time"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

func (f *TarFile) Getattr(ctx context.Context, fh fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	if filepath.Base(f.FilePath) == f.vfs.MainFile {
		out.Attr.Mode = uint32(0744) | fuse.S_IFREG
	} else {
		out.Attr.Mode = uint32(0644) | fuse.S_IFREG
	}
	out.Attr.Size = uint64(f.Header.Size)
	out.Size = out.Attr.Size

	out.SetTimeout(1 * time.Second)
	return 0
}

// Open unpacks tar
func (f *TarFile) Open(ctx context.Context, flags uint32) (fs.FileHandle, uint32, syscall.Errno) {
	f.mu.Lock()
	defer f.mu.Unlock()

	ino := f.Inode.StableAttr().Ino
	if f.vfs.FilesMap.Exist(ino) {
		f.Content = f.vfs.FilesMap.Get(ino).Content
	}

	// We don't return a filehandle since we don't really need
	// one.  The file content is immutable, so hint the kernel to
	// cache the data.
	return nil, fuse.FOPEN_KEEP_CACHE, fs.OK
}

// Read simply returns the data that was already unpacked in the Open call
func (f *TarFile) Read(ctx context.Context, fh fs.FileHandle, dest []byte, off int64) (fuse.ReadResult, syscall.Errno) {
	end := int(off) + len(dest)
	if end > len(f.Content) {
		end = len(f.Content)
	}
	return fuse.ReadResultData(f.Content[off:end]), fs.OK
}
