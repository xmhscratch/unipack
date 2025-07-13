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
	"strconv"
	"strings"
	"syscall"
	"time"
	"unipack/pkg/driver"

	"context"
	"path/filepath"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

func (vfs *VFSRoot) OnAdd(ctx context.Context) {
	// OnAdd is called once we are attached to an Inode. We can
	// then construct a tree.  We construct the entire tree, and
	// we don't want parts of the tree to disappear when the
	// kernel is short on memory, so we use persistent inodes.
	vfs.Walk(func(h *tar.Header, r *tar.Reader) error {
		dir, base := filepath.Split(h.Name)

		// fmt.Println(dir, base)
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
			} else {
				fi.Content = vfs.FilesMap.Get(ino).Content
			}
			vfs.FilesMap.Set(ino, fi)

			tarFile := &TarFile{FileInode: fi, vfs: vfs}
			ch = p.NewPersistentInode(ctx, tarFile, fs.StableAttr{Mode: fuse.S_IFREG})
		}

		relPath := p.Path(&vfs.Inode)
		if relPath == "." {
			vfs.Inode.AddChild(base, ch, true)
		} else {
			p.AddChild(base, ch, true)
		}
		return nil
	})

	if vfs.IsNew {
		vfs.FilesMap.Save()
	}
	vfs.readyMount <- struct{}{}
}

func (vfs *VFSRoot) Init() (err error) {
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

	vfs.HashSum = strconv.FormatUint(hashSum, 10)
	mntDir, err := vfs.createMountPoint()
	if err != nil {
		panic(err)
	}
	vfs.MountPoint = mntDir
	vfs.FilesMap = &FilesMap{uint64: hashSum, TFilesMap: &TFilesMap{}}
	vfs.IsNew, err = vfs.FilesMap.Load()

	vfs.readyMount = make(chan struct{}, 1)
	return err
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

func (vfs *VFSRoot) Serve(cleanCh chan struct{}) {
	var c chan os.Signal = make(chan os.Signal, 1)

	go func() {
		signal.Notify(c, os.Interrupt)

		var (
			srv       *fuse.Server
			isMounted bool          = false
			timeout   time.Duration = 30 * time.Second
		)
		if err := vfs.Init(); err != nil {
			panic(err)
		}
		srv, err := fs.Mount(vfs.MountPoint, vfs, &fs.Options{
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

		var errCh chan error = make(chan error)
		for {
			select {
			case <-vfs.readyMount:
				srv.WaitMount()
				isMounted = true
				go srv.Wait()
				go vfs.ExecBin(c, errCh)
			case <-time.After(30 * time.Second):
				c <- syscall.SIGCHLD
			case err := <-errCh:
				fmt.Println(err)
				c <- syscall.SIGHUP
			case sig := <-c:
				if isMounted {
					if err := srv.Unmount(); err != nil {
						panic(err)
					}
					if err := os.Remove(vfs.MountPoint); err != nil {
						panic(err)
					}
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
			}
		}
	}()
}

func (vfs *VFSRoot) ExecBin(sigCh chan os.Signal, errCh chan error) {
	ctx, cancel := context.WithCancelCause(context.Background())

	execFile := filepath.Join(vfs.MountPoint, vfs.MainFile)
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
		mntDir := filepath.Join(os.TempDir(), vfs.HashSum)
		if _, err := os.Stat(mntDir); os.IsNotExist(err) {
			if err = os.Mkdir(mntDir, os.ModeDir); err != nil {
				return mntDir, err
			}
		}
		return mntDir, err
	}

	if vfs.MountPoint != "" {
		mntDir, err := filepath.Abs(vfs.MountPoint)
		if err != nil {
			return makeDefaultMountPoint(err)
		}
		if _, err := os.Stat(mntDir); os.IsNotExist(err) {
			if err = os.Mkdir(mntDir, os.ModeDir); err != nil {
				return makeDefaultMountPoint(err)
			}
		}
	}
	return makeDefaultMountPoint(nil)
}
