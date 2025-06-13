package utils

import (
	"mime/multipart"
	"path/filepath"
	"strings"
)

type FileType string

const (
	PDF     FileType = "pdf"
	Archive FileType = "archive"
	Unknown FileType = "unknown"
)

func DetectFileType(fileHeader *multipart.FileHeader) FileType {
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))

	switch ext {
	case ".pdf":
		return PDF
	case ".zip", ".tar", ".gz", ".rar", ".7z":
		return Archive
	default:
		return Unknown
	}
}
