package filevalidation_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/mrbayss/golang-simple-ecommerce/internal/pkg/filevalidation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateImage(t *testing.T) {
	tests := []struct {
		name        string
		size        int64
		filename    string
		sniffedMIME string
		expectError bool
		errMsg      string
	}{
		{
			name:        "Valid JPEG image (.jpg)",
			size:        2 << 20, // 2 MB
			filename:    "product.jpg",
			sniffedMIME: "image/jpeg",
			expectError: false,
		},
		{
			name:        "Valid JPEG image (.jpeg uppercase)",
			size:        1 << 20,
			filename:    "photo.JPEG",
			sniffedMIME: "image/jpeg",
			expectError: false,
		},
		{
			name:        "Valid PNG image",
			size:        3 << 20,
			filename:    "banner.png",
			sniffedMIME: "image/png",
			expectError: false,
		},
		{
			name:        "Valid WebP image",
			size:        500 << 10,
			filename:    "thumbnail.webp",
			sniffedMIME: "image/webp",
			expectError: false,
		},
		{
			name:        "File exceeds max size (5 MB limit)",
			size:        6 << 20, // 6 MB
			filename:    "large.png",
			sniffedMIME: "image/png",
			expectError: true,
			errMsg:      "file exceeds the maximum size of 5 MB",
		},
		{
			name:        "Disallowed extension (.gif)",
			size:        1 << 20,
			filename:    "animation.gif",
			sniffedMIME: "image/gif",
			expectError: true,
			errMsg:      "file extension \".gif\" is not allowed",
		},
		{
			name:        "Disallowed executable disguised as image (.exe)",
			size:        1 << 20,
			filename:    "script.exe",
			sniffedMIME: "application/octet-stream",
			expectError: true,
			errMsg:      "file extension \".exe\" is not allowed",
		},
		{
			name:        "MIME type spoofing (extension .png but content is image/jpeg)",
			size:        1 << 20,
			filename:    "fake.png",
			sniffedMIME: "image/jpeg",
			expectError: true,
			errMsg:      "does not match its extension",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := filevalidation.ValidateImage(tt.size, tt.filename, tt.sniffedMIME)
			if tt.expectError {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestGenerateName(t *testing.T) {
	filename := "my_product_PHOTO.JPG"
	generated := filevalidation.GenerateName(filename)

	ext := filepath.Ext(generated)
	assert.Equal(t, ".jpg", ext)

	// Base name without extension must be a valid UUID
	base := strings.TrimSuffix(generated, ext)
	_, err := uuid.Parse(base)
	assert.NoError(t, err, "generated name base should be a valid UUID")

	// Consecutive calls should generate different unique names
	generated2 := filevalidation.GenerateName(filename)
	assert.NotEqual(t, generated, generated2)
}
