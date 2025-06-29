package uni

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"
	"xmhscratch/unipack/pkg/driver"

	"context"
	"path/filepath"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

func (vfs *VFSRoot) OnAdd(ctx context.Context) {
	isCreated, err := vfs.InitFilesMap()
	if err != nil {
		panic(err)
	}

	// OnAdd is called once we are attached to an Inode. We can
	// then construct a tree.  We construct the entire tree, and
	// we don't want parts of the tree to disappear when the
	// kernel is short on memory, so we use persistent inodes.
	vfs.Walk(func(h *tar.Header, r *tar.Reader) error {
		dir, base := filepath.Split(h.Name)

		p := &vfs.Inode
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

		if h.FileInfo().IsDir() {
			return nil
		}

		var (
			ch  *fs.Inode
			ino uint64     = p.StableAttr().Ino
			fi  *FileInode = &FileInode{
				Header:    h,
				MountPath: vfs.MountPoint,
				FilePath:  filepath.Join(p.Path(&vfs.Inode), base),
			}
		)

		{
			if !vfs.FilesMap.Exist(ino) {
				var content *bytes.Buffer = bytes.NewBuffer(make([]byte, 0))
				if _, err := io.CopyBuffer(content, r, make([]byte, 1024)); err != nil {
					panic(err)
				}
				fi.Content = content.Bytes()
			}

			tarFile := &TarFile{FileInode: fi, vfs: vfs}
			ch = p.NewPersistentInode(ctx, tarFile, fs.StableAttr{Mode: fuse.S_IFREG})
			vfs.FilesMap.Set(ino, fi)
		}

		relPath := p.Path(&vfs.Inode)
		if relPath == "." {
			vfs.Inode.AddChild(base, ch, true)
		} else {
			p.AddChild(base, ch, true)
		}
		return nil
	})

	if isCreated {
		vfs.FilesMap.Save()
	}
}

func (vfs *VFSRoot) Walk(walkFn func(*tar.Header, *tar.Reader) error) (err error) {
	var gzipReader *gzip.Reader
	{
		file, err := os.Open(vfs.TarFile)
		if err != nil {
			panic(err)
		}
		defer file.Close()

		gzf, err := gzip.NewReader(file)
		if err != nil {
			panic(err)
		}
		gzipReader = gzf
	}
	defer gzipReader.Close()

	var tarReader *tar.Reader = tar.NewReader(gzipReader)
	for {
		tarHeader, e := tarReader.Next()
		if e == io.EOF {
			break
		}
		if e != nil {
			err = e
			break
		}
		if e := walkFn(tarHeader, tarReader); e != nil {
			err = e
			break
		} else {
			continue
		}
	}

	return err
}

func (vfs *VFSRoot) InitFilesMap() (needSave bool, err error) {
	var gzipReader *gzip.Reader
	{
		var (
			file *os.File
			err  error
		)
		file, err = os.Open(vfs.TarFile)
		if err != nil {
			panic(err)
		}
		defer file.Close()

		gzf, err := gzip.NewReader(file)
		if err != nil {
			panic(err)
		}
		gzipReader = gzf
	}
	defer gzipReader.Close()

	var hashSum uint64
	{
		hash := crc32.New(crc32.MakeTable(crc32.Castagnoli))
		_, err := io.CopyBuffer(hash, gzipReader, make([]byte, 1024))
		if err != nil {
			panic(err)
		}
		hashSum = uint64(hash.Sum32())
	}

	vfs.FilesMap = &FilesMap{hashSum, TFilesMap{}}
	return vfs.FilesMap.Populate()
}

func (vfs *VFSRoot) InstantiateServer(cleanCh chan struct{}) {
	c := make(chan os.Signal, 1)

	mntDir, err := vfs.createMountPoint()
	if err != nil {
		panic(err)
	}

	var timeout time.Duration = 30 * time.Second
	srv, err := fs.Mount(mntDir, vfs, &fs.Options{
		EntryTimeout:    &timeout,
		AttrTimeout:     &timeout,
		NegativeTimeout: &timeout,
		NullPermissions: true,
		MountOptions: fuse.MountOptions{
			AllowOther: true, // allow non-root users
			Options: []string{
				"default_permissions", // enforce mode bits
				"exec",                // explicitly allow exec
			},
			Debug: false,
		},
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("tar file mounted at %s\n", vfs.MountPoint)

	go func() {
		signal.Notify(c, os.Interrupt)

		go func() {
			sig := <-c

			if err := srv.Unmount(); err != nil {
				panic(err)
			}
			if err := os.Remove(vfs.MountPoint); err != nil {
				panic(err)
			}
			cleanCh <- struct{}{}

			fmt.Println("exit signal: " + sig.String())
			switch sig {
			case syscall.SIGTERM:
				os.Exit(int(syscall.SIGTERM))
			case syscall.SIGINT:
				os.Exit(int(syscall.SIGINT))
			case syscall.SIGHUP:
				os.Exit(int(syscall.SIGHUP))
			case syscall.SIGKILL:
				os.Exit(int(syscall.SIGKILL))
			default:
				os.Exit(0)
			}
		}()

		// var errCh chan error = make(chan error)
		// vfs.ExecBin(c, errCh)
		// // fmt.Println(<-errCh)

		srv.Wait()
	}()
}

func (vfs *VFSRoot) ExecBin(sigCh chan os.Signal, errCh chan error) {
	execFile := filepath.Join(vfs.MountPoint, vfs.MainFile)
	// fmt.Println(execFile)

	ctx, cancel := context.WithCancelCause(context.Background())
	// cmd := exec.CommandContext(ctx, "factor 99")
	cmd := exec.CommandContext(ctx, execFile)
	cmd.Dir = vfs.MountPoint
	cmd.Env = os.Environ()

	go func() {
		select {
		case <-ctx.Done():
			errCh <- nil
			sigCh <- syscall.SIGHUP
			return
		case errCh <- ctx.Err():
			if <-errCh == nil {
				sigCh <- syscall.SIGTERM
			} else {
				sigCh <- syscall.SIGINT
			}
			return
		}
	}()

	var (
		err    error
		stdout io.ReadCloser
		stderr io.ReadCloser
	)
	if stdout, err = cmd.StdoutPipe(); err != nil {
		cancel(err)
	}
	defer stdout.Close()

	if stderr, err = cmd.StderrPipe(); err != nil {
		cancel(err)
	}
	defer stderr.Close()

	out := &driver.SocketStdout{}
	defer out.Close()

	if err := cmd.Start(); err != nil {
		cancel(err)
	}

	if _, err := io.CopyBuffer(out, stdout, make([]byte, 1024)); err != nil {
		cancel(err)
	}
	if _, err := io.CopyBuffer(out, stderr, make([]byte, 1024)); err != nil {
		cancel(err)
	}

	if err := cmd.Wait(); err != nil {
		cancel(err)
	} else {
		cancel(nil)
	}
}

func (vfs *VFSRoot) createMountPoint() (string, error) {
	makeDefaultMountPoint := func(err error) (string, error) {
		if err != nil {
			fmt.Printf("%s", err.Error())
		}

		mntDir, err := os.MkdirTemp("", "")
		vfs.MountPoint = mntDir

		return mntDir, err
	}
	// if vfs.MountPoint != "" {
	// 	fmt.Println(vfs.MountPoint)
	// 	// mntDir, err := filepath.Abs(vfs.MountPoint)
	// 	// if err != nil {
	// 	// 	return makeDefaultMountPoint(err)
	// 	// }
	// 	// err = os.Mkdir(mntDir, os.ModeDir)
	// 	// if err != nil {
	// 	// 	return makeDefaultMountPoint(err)
	// 	// }
	// }
	return makeDefaultMountPoint(nil)
}
