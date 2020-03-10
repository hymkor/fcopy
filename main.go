package main

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"time"

	_ "github.com/mattn/getwild"

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

func mains(args []string) error {
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
	} else {
		if len(args) != 2 {
			return fmt.Errorf("target '%s' is not a directory", dst)
		}
		copy1(args[0], args[1])
	}
	return nil
}

func main() {
	if err := mains(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
