package filevalidate

import (
	"fmt"
	"path/filepath"
	"strings"
)

var allowedExts = map[string]bool{
	".pdf":  true,
	".doc":  true,
	".docx": true,
}

func ValidateFilename(filename string) error {
	if filename == "" {
		return fmt.Errorf("filename is required")
	}
	if strings.Contains(filename, "..") || strings.Contains(filename, "/") || strings.Contains(filename, "\\") {
		return fmt.Errorf("invalid filename")
	}
	ext := strings.ToLower(filepath.Ext(filename))
	if !allowedExts[ext] {
		return fmt.Errorf("only PDF, DOC, DOCX are allowed")
	}
	return nil
}

func ValidateOSSKey(key string) error {
	if key == "" {
		return fmt.Errorf("oss key is required")
	}
	if strings.Contains(key, "..") {
		return fmt.Errorf("invalid oss key")
	}
	ext := strings.ToLower(filepath.Ext(key))
	if !allowedExts[ext] {
		return fmt.Errorf("only PDF, DOC, DOCX are allowed")
	}
	if !strings.HasPrefix(key, "resumes/") {
		return fmt.Errorf("invalid resume path")
	}
	return nil
}

func ValidateContentType(contentType string) error {
	allowed := map[string]bool{
		"application/pdf": true,
		"application/msword": true,
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
		"application/octet-stream": true,
	}
	if !allowed[strings.ToLower(contentType)] {
		return fmt.Errorf("invalid content type")
	}
	return nil
}
