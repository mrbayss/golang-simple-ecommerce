package storage

import (
	"context"
	"mime/multipart"
)

// FileStorage persists uploaded files and returns their public URL path.
type FileStorage interface {
	// Save stores the file under the given domain folder (e.g. "product")
	// and returns the public URL of the stored file.
	Save(ctx context.Context, domain string, file *multipart.FileHeader) (string, error)
	// Delete removes a previously stored file by its public URL path.
	Delete(ctx context.Context, url string) error
}
