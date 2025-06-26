package uni

import (
	"io"
	"os"
	"time"

	"context"
	"syscall"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

func (f *TarFile) Getattr(ctx context.Context, fh fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	mode := f.Mode()
	if mode&0111 != 0 { // check if any exec bit is set
		out.Attr.Mode = uint32(0755) | fuse.S_IFREG
	} else {
		out.Attr.Mode = uint32(0444) | fuse.S_IFREG
	}
	out.Attr.Size = uint64(f.Header.Size)
	out.Size = out.Attr.Size

	out.SetTimeout(1 * time.Second)
	return 0
}

// Open lazily unpacks tar data
func (f *TarFile) Open(ctx context.Context, flags uint32) (fs.FileHandle, uint32, syscall.Errno) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.data == nil {
		var (
			file *os.File
			err  error
		)
		file, err = os.Open(f.FilePath)
		if err != nil {
			panic(err)
		}
		defer file.Close()

		content, err := io.ReadAll(file)
		if err != nil {
			return nil, 0, syscall.EIO
		}
		f.data = content
	}

	// We don't return a filehandle since we don't really need
	// one.  The file content is immutable, so hint the kernel to
	// cache the data.
	return nil, fuse.FOPEN_KEEP_CACHE, fs.OK
}

// Read simply returns the data that was already unpacked in the Open call
func (f *TarFile) Read(ctx context.Context, fh fs.FileHandle, dest []byte, off int64) (fuse.ReadResult, syscall.Errno) {
	end := int(off) + len(dest)
	if end > len(f.data) {
		end = len(f.data)
	}
	return fuse.ReadResultData(f.data[off:end]), fs.OK
}
