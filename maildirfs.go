package main

// fs/serve.go says: "Operations where the nodes may return 0 inodes include Getattr, Setattr and ReadDir."

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	_ "bazil.org/fuse/fs/fstestutil"
)

func main() {

	log.SetFlags(0)

	// don't run as root (increases attack surface)

	if os.Geteuid() == 0 {
		log.Fatalln("refusing to run as root")
	}

	// Golang's flag package expects args after flags. Fuse does the other way round. Currently we don't use flags anyway.

	if len(os.Args) < 3 {
		log.Fatalln("usage: maildirfs source mountpoint")
	}

	var source = os.Args[1]
	var mountpoint = os.Args[2]

	// when running "mount /mountpoint" (while having an entry in fstab), PATH was not available and fuse.Mount failed to execute fusermount

	if os.Getenv("PATH") == "" {
		os.Setenv("PATH", "/usr/bin")
	}

	// unmount (executes "fusermount -u mountpoint") in case anything had crashed before

	if err := fuse.Unmount(mountpoint); err == nil {
		log.Println("unmounted mountpoint")
	} else {
		if !strings.Contains(err.Error(), "not found in /etc/mtab") {
			log.Fatalf("error unmounting mountpoint: %v", err)
		}
	}

	// FUSE mount

	conn, err := fuse.Mount(
		mountpoint,
		fuse.FSName(filepath.Base(mountpoint)),
		fuse.ReadOnly(),
		fuse.Subtype("maildirfs"),
	)
	if err != nil {
		log.Fatalf("error mounting: %v", err)
	}
	defer conn.Close()

	// Maildir++ root

	root, err := NewRoot(source)
	if err != nil {
		log.Printf("error creating root: %v", err)
		return
	}

	// FUSE server
	//
	// https://github.com/bazil/fuse/issues/6
	// "filesystem shutdown should be initiated by unmounting it, with fusermount -u."
	//
	// https://github.com/bazil/fuse/issues/214
	// "The Unmount function is only used for triggering the fusermount from inside the FUSE server process. That's pretty rare. It mostly only happens with unit tests."
	//
	// https://github.com/bazil/fuse/issues/244
	// "Serve will return when the kernel closes the FUSE connection."

	err = fs.Serve(conn, FS{root})
	if err != nil {
		log.Printf("error serving: %v", err)
	}
}

type FS struct {
	root *Root
}

func (fs FS) Root() (fs.Node, error) {
	return fs.root, nil
}
