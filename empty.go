package main

import (
	"context"
	"os"
	"syscall"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
)

type EmptyDir struct{}

func (e EmptyDir) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Mode = os.ModeDir | 0500
	return nil
}

func (e EmptyDir) Lookup(ctx context.Context, name string) (fs.Node, error) {
	return nil, syscall.ENOENT
}

func (e EmptyDir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	return nil, nil
}
