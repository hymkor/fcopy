package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	_ "github.com/mattn/getwild"

	"github.com/nyaosorg/go-windows-netresource"
	"github.com/nyaosorg/go-windows-su"

	"github.com/zetamatta/fcopy/internal/file"
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

func areSameFiles(path1, path2 string) bool {
	fd1, err := os.Open(path1)
	if err != nil {
		return false
	}
	defer fd1.Close()
	fd2, err := os.Open(path2)
	if err != nil {
		return false
	}
	defer fd2.Close()

	stat1, err := fd1.Stat()
	if err != nil {
		return false
	}
	stat2, err := fd2.Stat()
	if err != nil {
		return false
	}
	if stat1.Size() != stat2.Size() {
		return false
	}

	reader1 := bufio.NewReader(fd1)
	reader2 := bufio.NewReader(fd2)

	for {
		b1, err1 := reader1.ReadByte()
		b2, err2 := reader2.ReadByte()
		if err1 != nil {
			if err2 != nil {
				return true
			}
			return false
		} else if err2 != nil {
			return false
		}
		if b1 != b2 {
			return false
		}
	}
}

func tryCopy(src, dst string) error {
	if areSameFiles(src, dst) {
		fmt.Fprintf(os.Stderr, "'%s' and '%s' are same files.\n", src, dst)
		return nil
	}
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

func tryCopyFiles(args []string) error {
	dst := args[len(args)-1]
	status, err := isDir(dst)
	if err != nil {
		return err
	}
	if status == DirExist {
		for _, srcpath := range args[:len(args)-1] {
			name := filepath.Base(srcpath)
			dstpath := filepath.Join(dst, name)
			if err := tryCopy(srcpath, dstpath); err != nil {
				return err
			}
		}
		return nil
	} else {
		if len(args) != 2 {
			return fmt.Errorf("target '%s' is not a directory", dst)
		}
		return tryCopy(args[0], args[1])
	}
}

var flagPause = flag.Bool("pause", false, "pause after copy")

func mains(args []string) error {
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "fcopy files... dir")
		return nil
	}
	err := tryCopyFiles(args)
	if !isAccessDenied(err) {
		if *flagPause {
			fmt.Fprint(os.Stderr, "\n[Hit ENTER key]\n")
			var dummy [10]byte
			os.Stdin.Read(dummy[:])
		}
		return err
	}
	if *flagPause {
		return errors.New("To elevate with -flagPause is forbidden")
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
	buffer.WriteString(`/s /c "`)

	if netDrives, err := netresource.GetNetDrives(); err == nil {
		for _, n := range netDrives {
			fmt.Fprintf(&buffer, `net use %c: "%s" 2>nul & `,
				n.Letter, n.Remote)
		}
	}
	fmt.Fprintf(&buffer, `cd /d "%s" & "%s" -pause`, dir, me)

	for _, s := range args {
		fmt.Fprintf(&buffer, ` "%s"`, s)
	}
	buffer.WriteString(` "`)
	param := buffer.String()
	fmt.Println(param)

	_, err = su.ShellExecute(su.RUNAS, "CMD.EXE", param, "")
	return err
}

func main() {
	flag.Parse()
	if err := mains(flag.Args()); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
