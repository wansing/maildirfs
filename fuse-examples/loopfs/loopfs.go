// Simple FUSE example which mounts a folder to a mountpoint.
package main

import (
	"context"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"syscall"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
)

var source string
var mountpoint string

func main() {

	log.SetFlags(0)

	if len(os.Args) == 3 {
		source = os.Args[1]
		mountpoint = os.Args[2]
	} else {
		log.Fatalln("usage: %s source mountpoint", os.Args[0])
	}

	conn, err := fuse.Mount(
		mountpoint,
		fuse.FSName(filepath.Base(mountpoint)),
		fuse.ReadOnly(),
		fuse.Subtype("loopfs"),
	)
	if err != nil {
		log.Fatalln(err)
	}
	defer conn.Close()

	err = fs.Serve(conn, FS{})
	if err != nil {
		log.Println(err)
	}
}

type FS struct{}

func (FS) Root() (fs.Node, error) {
	return Dir(source), nil
}

// absolute source path
type Dir string

func (d Dir) Attr(ctx context.Context, a *fuse.Attr) error {
	info, err := os.Lstat(string(d))
	if err != nil {
		return err
	}
	a.Mode = info.Mode()
	a.Size = uint64(info.Size())
	return nil
}

func (d Dir) Lookup(ctx context.Context, name string) (fs.Node, error) {
	var path = filepath.Join(string(d), name)
	var info, err = os.Lstat(path)
	switch {
	case os.IsNotExist(err):
		return nil, syscall.ENOENT
	case err != nil:
		return nil, err
	case info.IsDir():
		return Dir(path), nil
	default:
		return File(path), nil
	}
}

func (d Dir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	f, err := os.Open(string(d))
	if err != nil {
		return nil, err
	}
	fileInfos, err := f.Readdir(0)
	if err != nil {
		return nil, err
	}
	dirents := []fuse.Dirent{}
	for _, fileInfo := range fileInfos {
		dirent := fuse.Dirent{
			// Inode: "Operations where the nodes may return 0 inodes include Getattr, Setattr and ReadDir."
			Name: fileInfo.Name(),
		}
		if fileInfo.IsDir() {
			dirent.Type = fuse.DT_Dir
		} else {
			dirent.Type = fuse.DT_File
		}
		dirents = append(dirents, dirent)
	}
	return dirents, nil
}

// absolute souce path
type File string

func (f File) Attr(ctx context.Context, a *fuse.Attr) error {
	info, err := os.Lstat(string(f))
	if err != nil {
		return err
	}
	a.Mode = info.Mode()
	a.Size = uint64(info.Size())
	return nil
}

func (f File) ReadAll(ctx context.Context) ([]byte, error) {
	return ioutil.ReadFile(string(f))
}
