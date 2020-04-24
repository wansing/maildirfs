# maildirfs

A FUSE filesystem which serves a folder as a Maildir.

## Try it

* build: `go build`
* start: `./maildirfs source mountpoint`
* stop: `fusermount -u mountpoint`

## Example

```
$ tree source/
source/
└── foo
    └── bar
        ├── A Document.pdf
        └── Hello World.md

$ tree -a mountpoint/
mountpoint/
├── cur
├── dovecot-acl
├── .foo
│   ├── cur
│   ├── dovecot-acl
│   ├── new
│   └── tmp
├── .foo.bar
│   ├── cur
│   │   ├── 1587735520.Hello World.md.localhost:2,S
│   │   └── 1587735553.A Document.pdf.localhost:2,S
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

## TODO

* add integration steps with overlayfs for index files etc
