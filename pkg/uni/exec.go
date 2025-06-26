package uni

import (
	"log"
	"syscall"
)

// "golang.org/x/net/context"

func ExecBin() {
	binaryPath := "/path/to/mybinary"
	args := []string{binaryPath, "arg1", "arg2"} // first arg must be program name
	env := []string{"PATH=/usr/bin:/bin"}

	err := syscall.Exec(binaryPath, args, env)
	if err != nil {
		log.Fatal("Exec failed:", err)
	}
}
