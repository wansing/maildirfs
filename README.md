# maildirfs

A FUSE filesystem which serves a directory and its subdirectories as a Maildir directory.

## Try it

* start: `./maildirfs source mountpoint`
* stop: `fusermount -u mountpoint`

## Example

```
$ tree source/
source/
├── bar
│   └── Text.txt
└── foo
    ├── A Document.pdf
    └── Hello World.md

$ tree -a mountpoint/
mountpoint/
├── .bar
│   ├── cur
│   │   └── 1587719883.Text.txt.localhost:2,S
│   ├── dovecot-acl
│   ├── new
│   └── tmp
├── cur
├── dovecot-acl
├── .foo
│   ├── cur
│   │   ├── 1587719860.Hello World.md.localhost:2,S
│   │   └── 1587719953.A Document.pdf.localhost:2,S
│   ├── dovecot-acl
│   ├── new
│   └── tmp
├── new
└── tmp
```

## Integration

* `maildirfs` executable must be in `$PATH`
* manual mount: `mount.fuse /path/to/source /path/to/mountpoint -t maildirfs`
* automatic mount: add to `/etc/fstab`: `/path/to/source /path/to/mountpoint fuse.maildirfs users`

## Design

* refuses to run as root
* folders in mountpoint have chmod `500`
* "maildirfolder" file has been omitted because the [Maildir spec](http://www.courier-mta.org/maildir.html) refers to it in context of a mail delivery agent only
