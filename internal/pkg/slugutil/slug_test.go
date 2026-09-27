package slugutil_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/mrbayss/golang-simple-ecommerce/internal/pkg/slugutil"
	"github.com/stretchr/testify/assert"
)

func TestSlugify(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Standard title with exclamation",
			input:    "Bakso Pedas Spesial!",
			expected: "bakso-pedas-spesial",
		},
		{
			name:     "Title with multiple spaces and special characters",
			input:    "  Kemeja Pria  -- Katun 100%   Original!  ",
			expected: "kemeja-pria-katun-100-original",
		},
		{
			name:     "Already slugified",
			input:    "sepatu-sneakers-keren",
			expected: "sepatu-sneakers-keren",
		},
		{
			name:     "Mixed case with underscores and symbols",
			input:    "iPhone_15_Pro_Max!@#$%^&*()",
			expected: "iphone-15-pro-max",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Only special characters",
			input:    "***---???",
			expected: "",
		},
		{
			name:     "Numbers only",
			input:    "12345",
			expected: "12345",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := slugutil.Slugify(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSlugify_Concurrency(t *testing.T) {
	const goroutines = 100
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			input := fmt.Sprintf("Produk Keren #%d Spesial!", id)
			expected := fmt.Sprintf("produk-keren-%d-spesial", id)
			result := slugutil.Slugify(input)
			assert.Equal(t, expected, result)
		}(i)
	}

	wg.Wait()
}
