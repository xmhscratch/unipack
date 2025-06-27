package uni

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"context"
	"path/filepath"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

func (f *VFSRoot) OnAdd(ctx context.Context) {
	// OnAdd is called once we are attached to an Inode. We can
	// then construct a tree.  We construct the entire tree, and
	// we don't want parts of the tree to disappear when the
	// kernel is short on memory, so we use persistent inodes.
	var (
		file *os.File
		err  error
	)
	file, err = os.Open(f.TarFile)
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
			Header:    tarHeader,
			MountPath: f.MountPoint,
			FilePath:  filepath.Join(p.Path(&f.Inode), base),
		}, fs.StableAttr{Mode: fuse.S_IFREG})

		relPath := p.Path(&f.Inode)
		if relPath == "." {
			f.Inode.AddChild(base, ch, true)
		} else {
			p.AddChild(base, ch, true)
		}
	}
}

func (f *VFSRoot) InstantiateServer(cleanCh chan struct{}) {
	c := make(chan os.Signal, 1)

	mntDir, err := f.createMountPoint()
	if err != nil {
		panic(err)
	}

	var timeout time.Duration = 5 * time.Minute
	srv, err := fs.Mount(mntDir, f, &fs.Options{
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
			Debug: true,
		},
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("tar file mounted at %s\n", f.MountPoint)

	if err := f.ExecBin(); err != nil {
		fmt.Println(err)
		c <- os.Interrupt
	}

	go func() {
		signal.Notify(c, os.Interrupt)

		go func() {
			sig := <-c

			if err := srv.Unmount(); err != nil {
				panic(err)
			}
			if err := os.Remove(f.MountPoint); err != nil {
				panic(err)
			}
			cleanCh <- struct{}{}

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
				fmt.Println(sig.String())
				os.Exit(0)
			}
		}()
		srv.Wait()
	}()
}

func (f *VFSRoot) ExecBin() error {
	// execFile := filepath.Join(f.MountPoint, f.MainFile)
	// fmt.Println(execFile)
	// cmd := exec.Command(execFile)

	// stdout, err := cmd.StdoutPipe()
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer stdout.Close()

	// out := &driver.SocketStdout{}
	// defer out.Close()

	// if err := cmd.Start(); err != nil {
	// 	log.Fatal(err)
	// }

	// _, err = io.CopyBuffer(out, stdout, copyBuffer)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// if err := cmd.Wait(); err != nil {
	// 	log.Fatal(err)
	// }
	return nil
}

func (f *VFSRoot) createMountPoint() (string, error) {
	makeDefaultMountPoint := func(err error) (string, error) {
		if err != nil {
			fmt.Printf("%s", err.Error())
		}

		mntDir, err := os.MkdirTemp("", "")
		f.MountPoint = mntDir

		return mntDir, err
	}
	// if f.MountPoint != "" {
	// 	fmt.Println(f.MountPoint)
	// 	// mntDir, err := filepath.Abs(f.MountPoint)
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
