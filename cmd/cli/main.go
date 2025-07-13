package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unipack/pkg/uni"
)

// StringMapFlag implements flag.Value for a map[string]string
type StringMapFlag map[string]string

// Set parses the string argument into the map
func (sm StringMapFlag) Set(value string) error {
	pairs := strings.Split(value, ",")
	for _, pair := range pairs {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) == 2 {
			sm[parts[0]] = parts[1]
		} else {
			return fmt.Errorf("invalid map format: %s", pair)
		}
	}
	return nil
}

// String returns the string representation of the map
func (sm StringMapFlag) String() string {
	var s []string
	for k, v := range sm {
		s = append(s, fmt.Sprintf("%s=%s", k, v))
	}
	return strings.Join(s, ",")
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("'run' commands not specified")
		os.Exit(1)
	}

	var cleanCh chan struct{} = make(chan struct{})
	// defer close(cleanCh)

	switch os.Args[1] {
	case "run":
		runCmd := flag.NewFlagSet("run", flag.ExitOnError)
		{
			var (
				mainFile   string
				mountPoint string
				entries    StringMapFlag
			)
			runCmd.StringVar(&mainFile, "main-file", "main", "application main execution binary file")
			runCmd.StringVar(&mountPoint, "mount-point", "", "(optional) define default mount point on the host machine")
			runCmd.Var(&entries, "entries", "(optional) mounting external files or folders at runtime")
			runCmd.Parse(os.Args[2:])

			// fmt.Println(mainFile, mountPoint, entries)
			// fmt.Println(runCmd.Arg(0))

			if runCmd.Arg(0) == "" {
				panic("usage: Tar-FILE not specified")
			}

			tarFile, err := filepath.Abs(runCmd.Arg(0))
			if err != nil {
				panic(err)
			}

			(&uni.VFSRoot{
				TarFile:    tarFile,
				MainFile:   mainFile,
				MountPoint: mountPoint,
			}).Serve(cleanCh)
		}
	default:
		os.Exit(1)
	}

	// wait for exit signal
	<-cleanCh
}
