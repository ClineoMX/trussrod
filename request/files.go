package request

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"

	"github.com/Domedik/trussrod/errors"
	"github.com/Domedik/trussrod/storage"
	"github.com/google/uuid"
)

func ingest(part *multipart.Part) (*storage.File, error) {
	defer part.Close()
	id := uuid.New()
	fileName := part.FileName()
	if fileName == "" || fileName == "." || fileName == ".." {
		return nil, fmt.Errorf("invalid file name")
	}

	var buf bytes.Buffer
	hasher := sha256.New()

	writer := io.MultiWriter(&buf, hasher)
	_, err := io.Copy(writer, part)
	if err != nil {
		return nil, err
	}
	file := &storage.File{
		ID:      id.String(),
		Content: buf.Bytes(),
		Hash:    hasher.Sum(nil),
		Name:    fileName,
	}
	return file, nil
}

func Files(r *http.Request) ([]*storage.File, error) {
	var files []*storage.File

	_, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil {
		return files, err
	}
	reader := multipart.NewReader(r.Body, params["boundary"])
	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return files, err
		}

		if part.FormName() == "files" {
			f, err := ingest(part)
			if err != nil {
				continue
			}
			files = append(files, f)
		} else {
			_, _ = io.Copy(io.Discard, part)
		}
	}

	if len(files) == 0 {
		return files, errors.BadRequest("no files sent")
	}

	return files, nil
}
