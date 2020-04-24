package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
)

// implements fs.Node
type Root struct {
	Dir         // absolute path
	cache       []fuse.Dirent
	cacheLocker sync.Mutex
	changed     int64
}

func NewRoot(source string) (*Root, error) {
	source, err := filepath.Abs(source)
	return &Root{
		Dir: Dir(source),
	}, err
}

// func (*Root) Attr is inherited from Dir

func (root *Root) Lookup(ctx context.Context, name string) (fs.Node, error) {
	var node, err = root.Dir.Lookup(ctx, name)
	if err != nil {
		node = Dir(filepath.Join(string(root.Dir), maildirToOS(name)))
	}
	return node, nil
}

func (root *Root) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {

	// must stat to get mod time
	info, err := os.Lstat(string(root.Dir))
	if err != nil {
		return nil, err
	}

	root.cacheLocker.Lock()

	if root.changed < info.ModTime().Unix() {
		root.cache = root.cache[:0]
		root.changed = info.ModTime().Unix()
		if err = root.appendDirs("."); err != nil {
			return nil, err
		}
	}

	root.cacheLocker.Unlock()

	var parentAll, _ = root.Dir.ReadDirAll(ctx)
	var all = append(root.cache, parentAll...)

	return all, nil
}

// own Walk implementation, probably faster than filepath.Walk but could be improved by selective caching

func (root *Root) appendDirs(relpath string) error {

	f, err := os.Open(filepath.Join(string(root.Dir), relpath))
	if err != nil {
		return err
	}

	contents, err := f.Readdir(0)
	if err != nil {
		return err
	}

	f.Close() // before recursing

	for _, content := range contents {
		var relChildPath = filepath.Join(relpath, content.Name())
		if content.IsDir() {
			root.cache = append(root.cache, fuse.Dirent{
				Name: osToMaildir(relChildPath),
				Type: fuse.DT_Dir,
			})
			if err := root.appendDirs(relChildPath); err != nil {
				return err
			}
		}
	}
	return nil
}

// returns like "/Foo/Bar"
func maildirToOS(name string) string {
	var path = strings.ReplaceAll(name, ".", string(os.PathSeparator))
	path = strings.ReplaceAll(path, `\.`, ".")
	return path
}

// returns like ".Foo.Bar"
func osToMaildir(path string) string {
	var name = strings.ReplaceAll(path, ".", `\.`)
	name = strings.ReplaceAll(name, string(os.PathSeparator), ".")
	if !strings.HasPrefix(name, ".") {
		name = "." + name
	}
	return name
}
