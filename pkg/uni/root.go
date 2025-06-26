package uni

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"strings"

	"context"
	"path/filepath"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

func (f *TarRoot) OnAdd(ctx context.Context) {
	// OnAdd is called once we are attached to an Inode. We can
	// then construct a tree.  We construct the entire tree, and
	// we don't want parts of the tree to disappear when the
	// kernel is short on memory, so we use persistent inodes.
	var (
		file *os.File
		err  error
	)
	file, err = os.Open(f.TarPath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	gzf, err := gzip.NewReader(file)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	tr := tar.NewReader(gzf)
	for {
		tarHeader, err := tr.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			panic(err)
			// break
		}
		dir, base := filepath.Split(tarHeader.Name)

		p := &f.Inode
		for _, component := range strings.Split(dir, "/") {
			if len(component) == 0 {
				continue
			}
			ch := p.GetChild(component)
			if ch == nil {
				ch = p.NewPersistentInode(ctx, &fs.Inode{}, fs.StableAttr{Mode: fuse.S_IFDIR})
				p.AddChild(component, ch, true)
			}
			p = ch
		}

		if tarHeader.FileInfo().IsDir() {
			continue
		}

		ch := p.NewPersistentInode(ctx, &TarFile{
			Header:   tarHeader,
			FilePath: filepath.Join(p.Path(&f.Inode), base),
		}, fs.StableAttr{Mode: fuse.S_IFREG})

		relPath := p.Path(&f.Inode)
		if relPath == "." {
			f.Inode.AddChild(base, ch, true)
		} else {
			p.AddChild(base, ch, true)
		}
	}
}
