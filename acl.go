package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"bazil.org/fuse"
)

var acls = map[DovecotACL]struct {
	Content []byte
	Changed int64
}{}
var aclsLocker sync.Mutex

// absolute filesystem path to Dir
type DovecotACL string

func (f DovecotACL) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Mode = 0400
	var acl, err = f.ReadAll(nil)
	if err != nil {
		return err
	}
	a.Size = uint64(len(acl))
	return nil
}

func (f DovecotACL) ReadAll(ctx context.Context) ([]byte, error) {

	// must stat to get mod time
	info, err := os.Lstat(filepath.Join(string(f), ".readers"))
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	aclsLocker.Lock()

	content, ok := acls[f]
	if !ok || content.Changed < info.ModTime().Unix() {
		content.Changed = info.ModTime().Unix()
		content.Content, err = f.read()
		if err != nil {
			return nil, err
		}
		acls[f] = content // write back
	}

	aclsLocker.Unlock()

	return content.Content, nil
}

func (f DovecotACL) read() ([]byte, error) {
	var readersFile, err = os.Open(filepath.Join(string(f), ".readers"))
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var acl = &bytes.Buffer{}
	var scanner = bufio.NewScanner(readersFile)
	for scanner.Scan() {
		if line := strings.TrimSpace(scanner.Text()); line != "" {
			acl.WriteString(fmt.Sprintf("user=%s lr\n", line))
		}
	}
	return acl.Bytes(), nil
}
