package main

import (
	"fmt"
	"os"
	"os/signal"
	"xmhscratch/unipack/pkg/uni"

	"flag"
	"log"
	"path/filepath"
	"syscall"

	"github.com/hanwen/go-fuse/v2/fs"
)

// ExampleTarFS shows an in-memory, static file system
func main() {
	flag.Parse()
	if len(flag.Args()) != 1 {
		log.Fatal("usage: tarmount Tar-FILE")
	}

	fileName, err := filepath.Abs(flag.Arg(0))
	if err != nil {
		panic(err)
	}
	root := &uni.TarRoot{TarPath: fileName}

	mntDir, _ := os.MkdirTemp("", "")
	server, err := fs.Mount(mntDir, root, nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("tar file mounted")
	fmt.Printf("to unmount: fusermount -u %s\n", mntDir)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		sig := <-c

		if err := server.Unmount(); err != nil {
			panic(err)
		}
		if err := os.Remove(mntDir); err != nil {
			panic(err)
		}

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
	server.Wait()
}
