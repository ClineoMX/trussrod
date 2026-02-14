package request

import (
	"mime/multipart"
	"net/http"
)

func Files(r *http.Request) ([]*multipart.FileHeader, error) {
	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		return nil, err
	}

	files := r.MultipartForm.File["files"]
	return files, nil
}
