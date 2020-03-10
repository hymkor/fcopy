package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	_ "github.com/mattn/getwild"

	"github.com/zetamatta/go-windows-su"

	"github.com/zetamatta/fcopy/file"
)

const (
	_ERRNO_USED_ANOTHER_PROCESS = 32
	_ERRNO_ACCESS_IS_DENIED     = 5
)

func isUsedAnotherProcess(err error) bool {
	e, ok := err.(syscall.Errno)
	return ok && e == _ERRNO_USED_ANOTHER_PROCESS
}

func isAccessDenied(err error) bool {
	e, ok := err.(syscall.Errno)
	return ok && e == _ERRNO_ACCESS_IS_DENIED
}

func copy1(src, dst string) error {
	err := file.Copy(src, dst, false)
	if err != nil {
		if !isUsedAnotherProcess(err) {
			return err
		}
		fmt.Fprintln(os.Stderr, err.Error())
		backup := dst + time.Now().Format("-20060102_150405")
		err = file.Move(dst, backup)
		if err != nil {
			return err
		}
		fmt.Printf("%s: renamed to %s\n", dst, backup)
		err = file.Copy(src, dst, false)
		if err != nil {
			return err
		}
	}
	fmt.Printf("%s -> %s\n", src, dst)
	return nil
}

type FileStatus int

const (
	PathNotFound FileStatus = iota
	DirExist
	FileExist
	DirOrFileError
)

func isDir(fname string) (FileStatus, error) {
	f, err := os.Stat(fname)
	if err != nil {
		if os.IsNotExist(err) {
			return PathNotFound, nil
		}
		return DirOrFileError, err
	}
	if f.IsDir() {
		return DirExist, nil
	} else {
		return FileExist, nil
	}
}

func main2(args []string) error {
	dst := args[len(args)-1]
	status, err := isDir(dst)
	if err != nil {
		return err
	}
	if status == DirExist {
		for _, srcpath := range args[:len(args)-1] {
			name := filepath.Base(srcpath)
			dstpath := filepath.Join(dst, name)
			if err := copy1(srcpath, dstpath); err != nil {
				return err
			}
		}
		return nil
	} else {
		if len(args) != 2 {
			return fmt.Errorf("target '%s' is not a directory", dst)
		}
		return copy1(args[0], args[1])
	}
}

var flagParse = flag.Bool("pause", false, "pause after copy")

func main1(args []string) error {
	err := main2(args)
	if !isAccessDenied(err) {
		if *flagParse {
			fmt.Fprint(os.Stderr, "\n[Hit ENTER key]\n")
			var dummy [10]byte
			os.Stdin.Read(dummy[:])
		}
		return err
	}
	me, err := os.Executable()
	if err != nil {
		return err
	}
	dir, err := os.Getwd()
	if err != nil {
		return err
	}

	var buffer strings.Builder
	buffer.WriteString("-pause")
	for _, s := range args {
		fmt.Fprintf(&buffer, ` "%s"`, s)
	}

	_, err = su.ShellExecute(su.RUNAS, me, buffer.String(), dir)
	return err
}

func main() {
	flag.Parse()
	if err := main1(flag.Args()); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
