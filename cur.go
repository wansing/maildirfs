package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
)

// absolute filesystem path to Dir
type CurDir string

func (c CurDir) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Mode = os.ModeDir | 0500
	return nil
}

func (c CurDir) Lookup(ctx context.Context, name string) (fs.Node, error) {
	name = msgToFile(name)
	var path = filepath.Join(string(c), name)
	var info, err = os.Lstat(path)
	switch {
	case os.IsNotExist(err):
		return nil, syscall.ENOENT
	case err != nil:
		return nil, err
	case info.IsDir():
		return nil, syscall.ENOENT // ignore dirctories, they are managed by root
	default:
		return File(path), nil
	}
}

func (c CurDir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	f, err := os.Open(string(c))
	if err != nil {
		return nil, err
	}
	fileInfos, err := f.Readdir(0)
	if err != nil {
		return nil, err
	}
	dirents := []fuse.Dirent{}
	for _, info := range fileInfos {
		if info.IsDir() { // the cur directory contains files only
			continue
		}
		dirents = append(
			dirents,
			fuse.Dirent{
				Name: fileToMsg(info),
				Type: fuse.DT_File,
			},
		)
	}

	return dirents, nil
}

func msgToFile(msgname string) string {
	if firstDot := strings.Index(msgname, "."); firstDot >= 0 {
		msgname = msgname[firstDot+1:]
	}
	if lastDot := strings.LastIndex(msgname, "."); lastDot >= 0 {
		msgname = msgname[:lastDot]
	}
	return msgname
}

// Maildir spec:
// "The name of the file in new should be [...] or 'time.MusecPpidVdevIino_unique.host,S=cnt'"
// "Rename new/filename, as cur/filename:2,info"
//
// https://cr.yp.to/proto/maildir.html
// "On the left is the result of time() or the second counter from gettimeofday(). On the right is the result of gethostname(). [...] In the middle is a delivery identifier"
// "When you move a file from new to cur, you have to change its name from uniq to uniq:info."
//
// real life example:
// 1445972235.M123739P14253.example.com,S=2917,W=2991:2,RSa
//
// dovecot https://wiki.dovecot.org/MailboxFormat/Maildir
// ",S=<size>: <size> contains the file size. Getting the size from the filename avoids doing a stat(), which may improve the performance."
//
// getting size from filesizes would save one Lstat
func fileToMsg(fileinfo os.FileInfo) string {
	return fmt.Sprintf("%d.%s.%s:2,S", fileinfo.ModTime().Unix(), fileinfo.Name(), "localhost")
}
