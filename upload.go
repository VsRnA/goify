package goify

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type FileHeader struct {
	*multipart.FileHeader
	File multipart.File
}

type FileValidation struct {
	MaxSize	int64
	MinSize	int64
	AllowedTypes []string
	AllowedExts	[]string
	Required bool
}

type FileUploadError	struct {
	Field string `json:"field"`
	Message string `json:"message"`
	Code	string `json:"code"`
}

func (fue FileUploadError) Error() string {
	return fue.Message
}

type FileUploadErrors []FileUploadError 

func (fues FileUploadErrors) Error() string {
	var messages []string

	for _, err := range fues {
		messages = append(messages, err.Message)
	}

	return strings.Join(messages, "; ")
}

func ValidateFile(fileHeader *FileHeader, validation FileValidation) error {
	if fileHeader == nil {
		if validation.Required {
			return FileUploadError{
				Message: "File is required",
				Code:    "required",
			}
		}
		return nil
	}

	if validation.MaxSize > 0 && fileHeader.Size > validation.MaxSize {
		return FileUploadError{
			Message: fmt.Sprintf("File size exceeds maximum allowed size of %d bytes", validation.MaxSize),
			Code:    "max_size",
		}
	}

	if validation.MinSize > 0 && fileHeader.Size < validation.MinSize {
		return FileUploadError{
			Message: fmt.Sprintf("File size is below minimum required size of %d bytes", validation.MinSize),
			Code:    "min_size",
		}
	}

	if len(validation.AllowedTypes) > 0 {
		allowed := false
		for _, allowedType := range validation.AllowedTypes {
			if fileHeader.Header.Get("Content-Type") == allowedType {
				allowed = true
				break
			}
		}
		if !allowed {
			return FileUploadError{
				Message: fmt.Sprintf("File type '%s' is not allowed. Allowed types: %v", 
					fileHeader.Header.Get("Content-Type"), validation.AllowedTypes),
				Code: "invalid_type",
			}
		}
	}

	if len(validation.AllowedExts) > 0 {
		ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
		allowed := false
		for _, allowedExt := range validation.AllowedExts {
			if ext == strings.ToLower(allowedExt) {
				allowed = true
				break
			}
		}
		if !allowed {
			return FileUploadError{
				Message: fmt.Sprintf("File extension '%s' is not allowed. Allowed extensions: %v", 
					ext, validation.AllowedExts),
				Code: "invalid_extension",
			}
		}
	}

	return nil
}

func SaveFile(fileHeader *FileHeader, dst string) error {
	if fileHeader == nil {
		return fmt.Errorf("file header is nil")
	}

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %v", err)
	}

	out, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %v", err)
	}
	defer out.Close()

	_, err = io.Copy(out, fileHeader.File)
	if err != nil {
		return fmt.Errorf("failed to copy file content: %v", err)
	}

	return nil
}

func SaveFileWithName(fileHeader *FileHeader, dir, filename string) error {
	dst := filepath.Join(dir, filename)
	return SaveFile(fileHeader, dst)
}

func GenerateUniqueFilename(originalName string) string {
	ext := filepath.Ext(originalName)
	base := strings.TrimSuffix(originalName, ext)

	base = strings.ReplaceAll(base, " ", "_")
	base = strings.ReplaceAll(base, "..", "")

	timestamp := fmt.Sprintf("%d", GetCurrentTimestamp())
	
	return fmt.Sprintf("%s_%s%s", base, timestamp, ext)
}

func GetCurrentTimestamp() int64 {
	return time.Now().UnixNano()
}

func GetFileSize(filename string) (int64, error) {
	info, err := os.Stat(filename)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

func DeleteFile(filename string) error {
	return os.Remove(filename)
}

func GetMimeType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	
	mimeTypes := map[string]string{
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".pdf":  "application/pdf",
		".txt":  "text/plain",
		".csv":  "text/csv",
		".json": "application/json",
		".xml":  "application/xml",
		".zip":  "application/zip",
		".mp4":  "video/mp4",
		".mp3":  "audio/mpeg",
	}
	
	if mimeType, exists := mimeTypes[ext]; exists {
		return mimeType
	}
	
	return "application/octet-stream"
}

func IsImageFile(mimeType string) bool {
	imageTypes := []string{
		"image/jpeg",
		"image/jpg", 
		"image/png",
		"image/gif",
		"image/webp",
		"image/svg+xml",
	}
	
	for _, imageType := range imageTypes {
		if mimeType == imageType {
			return true
		}
	}
	
	return false
}

func FormatFileSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	
	units := []string{"KB", "MB", "GB", "TB", "PB"}
	return fmt.Sprintf("%.1f %s", float64(size)/float64(div), units[exp])
}