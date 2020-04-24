package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"mime"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"text/template"
	"time"

	"bazil.org/fuse"
)

var filesizes = map[File]struct {
	Size    uint64
	Changed int64
}{}
var filesizesLocker sync.Mutex

// RFC 5322: Date and From fields are mandatory
//
// RFC 2183: "Content-Disposition" header field [is] valid for any MIME entity ("message" or "body part")
var mailTemplate = template.Must(template.New("").Parse(`Content-Disposition: {{ .ContentDisposition }}
Content-Transfer-Encoding: base64
Content-Type: {{ .ContentType }}
{{ with .Date -}}
Date: {{ . }}
{{ end -}}
From: {{ .From }}
Subject: {{ .Subject }}

{{ .Body }}`))

type mailTemplateData struct {
	body     []byte
	modTime  time.Time
	filepath string
}

func (data *mailTemplateData) Body() string {

	encoded := make([]byte, base64.StdEncoding.EncodedLen(len(data.body)))
	base64.StdEncoding.Encode(encoded, data.body)

	var buf = &bytes.Buffer{}

	for len(encoded) > 76 {
		buf.Write(encoded[:76])
		buf.WriteString("\n")
		encoded = encoded[76:]
	}
	buf.Write(encoded)

	return buf.String()
}

func (data *mailTemplateData) ContentDisposition() string {
	return fmt.Sprintf("inline")
}

func (data *mailTemplateData) ContentType() string {
	return mime.TypeByExtension(filepath.Ext(data.filepath))
}

func (data *mailTemplateData) Date() string {
	return data.modTime.Format(time.RFC822Z)
}

func (data *mailTemplateData) From() string {
	for _, segment := range strings.Split(data.filepath, string(os.PathSeparator)) {
		if strings.Contains(segment, "@") {
			return segment
		}
	}
	return "no-reply@example.com"
}

func (data *mailTemplateData) Subject() string {
	return filepath.Base(data.filepath)
}

type Inline struct {
	*mailTemplateData
}

type Html struct {
	*mailTemplateData
}

type Attachment struct {
	*mailTemplateData
}

func (a Attachment) ContentDisposition() string {
	return fmt.Sprintf("attachment; filename=%s", filepath.Base(a.filepath))
}

// absolute filesystem path to file
type File string

func (f File) Attr(ctx context.Context, a *fuse.Attr) error {

	// must stat to get mod time
	info, err := os.Lstat(string(f))
	if err != nil {
		return err
	}
	a.Mode = info.Mode()

	filesizesLocker.Lock()

	size, ok := filesizes[f]
	if !ok || size.Changed < info.ModTime().Unix() {
		var email, err = f.ReadAll(nil)
		if err != nil {
			return err
		}
		size.Changed = info.ModTime().Unix()
		size.Size = uint64(len(email))
		filesizes[f] = size // write back
	}

	filesizesLocker.Unlock()

	a.Size = size.Size
	return nil
}

func (f File) ReadAll(ctx context.Context) ([]byte, error) {

	var info, err = os.Lstat(string(f)) // need it for modification time
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadFile(string(f))
	if err != nil {
		return nil, err
	}

	var templateData = &mailTemplateData{
		body:     body,
		modTime:  info.ModTime(),
		filepath: string(f),
	}

	var data interface{}
	switch filepath.Ext(string(f)) {
	case ".md":
		data = Inline{templateData}
	case ".txt":
		data = Inline{templateData}
	case ".html":
		data = Html{templateData}
	default:
		data = Attachment{templateData}
	}

	var buf = &bytes.Buffer{}
	if err = mailTemplate.Execute(buf, data); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
