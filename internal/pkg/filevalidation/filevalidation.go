package filevalidation

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

var (
	allowedExts  = map[string]string{".jpg": "image/jpeg", ".jpeg": "image/jpeg", ".png": "image/png", ".webp": "image/webp"}
	maxFileSize  = int64(5 << 20) // 5 MB
	errSizeLimit = fmt.Errorf("file exceeds the maximum size of 5 MB")
)

// ValidateImage checks size, extension, and MIME type (sniffed from content,
// not from the filename) against the image allowlist.
func ValidateImage(size int64, filename string, sniffedMIME string) error {
	if size > maxFileSize {
		return errSizeLimit
	}

	ext := strings.ToLower(filepath.Ext(filename))
	expectedMIME, ok := allowedExts[ext]
	if !ok {
		return fmt.Errorf("file extension %q is not allowed", ext)
	}

	// Extension and real content must agree; a mismatch means the file is disguised.
	if sniffedMIME != expectedMIME {
		return fmt.Errorf("file content (%s) does not match its extension %q", sniffedMIME, ext)
	}

	return nil
}

// GenerateName replaces the original filename with a random UUID-based name,
// keeping only the lowercased extension. Prevents path traversal and collisions.
func GenerateName(originalFilename string) string {
	ext := strings.ToLower(filepath.Ext(originalFilename))
	return uuid.NewString() + ext
}
