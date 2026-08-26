package storage

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/mrbayss/golang-simple-ecommerce/internal/pkg/filevalidation"
)

// LocalStorage keeps files on disk under basePath/<domain>/ with
// random UUID filenames. Folders are created on demand.
type LocalStorage struct {
	basePath   string // e.g. "./public"
	publicPath string // e.g. "/public" — URL prefix served statically
}

func NewLocalStorage(basePath, publicPath string) *LocalStorage {
	return &LocalStorage{basePath: basePath, publicPath: publicPath}
}

// Save validates the file, stores it as basePath/<domain>/<uuid>.<ext>,
// and returns its public URL (publicPath + "/" + domain + "/" + filename).
func (ls *LocalStorage) Save(ctx context.Context, domain string, file *multipart.FileHeader) (string, error) {
	sniffedMIME, err := sniffMIME(file)
	if err != nil {
		return "", fmt.Errorf("read file content: %w", err)
	}

	if err := filevalidation.ValidateImage(file.Size, file.Filename, sniffedMIME); err != nil {
		return "", err
	}

	dir := filepath.Join(ls.basePath, domain)
	// MkdirAll is a no-op when the folder already exists.
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create storage dir: %w", err)
	}

	filename := filevalidation.GenerateName(file.Filename)

	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("open uploaded file: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(filepath.Join(dir, filename))
	if err != nil {
		return "", fmt.Errorf("create destination file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return "", fmt.Errorf("write file: %w", err)
	}

	url := ls.publicPath + "/" + domain + "/" + filename
	return url, nil
}

// Delete removes the file behind a stored URL. A missing file is not an
// error — the DB row is gone either way.
func (ls *LocalStorage) Delete(ctx context.Context, url string) error {
	if !strings.HasPrefix(url, ls.publicPath+"/") {
		return fmt.Errorf("url %q is outside storage scope", url)
	}

	rel := strings.TrimPrefix(url, ls.publicPath+"/")
	if strings.Contains(rel, "..") {
		return fmt.Errorf("invalid file path")
	}

	err := os.Remove(filepath.Join(ls.basePath, rel))
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove file: %w", err)
	}
	return nil
}

// sniffMIME detects the real content type from the first bytes of the file.
func sniffMIME(file *multipart.FileHeader) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	buf := make([]byte, 512)
	n, err := src.Read(buf)
	if err != nil && err != io.EOF {
		return "", err
	}

	return http.DetectContentType(buf[:n]), nil
}
