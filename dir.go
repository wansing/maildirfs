package main

import (
	"context"
	"os"
	"syscall"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
)

// absolute filesystem path to dir
type Dir string

func (md Dir) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Mode = os.ModeDir | 0500
	return nil
}

func (md Dir) Lookup(ctx context.Context, name string) (fs.Node, error) {
	switch name {
	case "cur":
		return CurDir(md), nil
	case "new":
		return EmptyDir{}, nil
	case "tmp":
		return EmptyDir{}, nil
	case "dovecot-acl":
		return DovecotACL(md), nil
	default:
		return nil, syscall.ENOENT
	}
}

func (md Dir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	return []fuse.Dirent{
		{Name: "cur", Type: fuse.DT_Dir},
		{Name: "new", Type: fuse.DT_Dir},
		{Name: "tmp", Type: fuse.DT_Dir},
		{Name: "dovecot-acl", Type: fuse.DT_File},
	}, nil
}
