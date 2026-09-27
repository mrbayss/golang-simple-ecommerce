package phonenumber_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/mrbayss/golang-simple-ecommerce/internal/pkg/phonenumber"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalize(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    string
		expectError bool
		errMsg      string
	}{
		{
			name:        "Valid starting with 08",
			input:       "08123456789",
			expected:    "628123456789",
			expectError: false,
		},
		{
			name:        "Valid starting with +62",
			input:       "+62 812 3456 789",
			expected:    "628123456789",
			expectError: false,
		},
		{
			name:        "Valid with dashes and spaces",
			input:       "0812-3456-7890",
			expected:    "6281234567890",
			expectError: false,
		},
		{
			name:        "Valid starting with 62 directly",
			input:       "628123456789",
			expected:    "628123456789",
			expectError: false,
		},
		{
			name:        "Invalid prefix (e.g. starts with 34)",
			input:       "34567890123",
			expectError: true,
			errMsg:      "phone number must start with 08 or 62",
		},
		{
			name:        "Invalid prefix (e.g. starts with 1)",
			input:       "12345678901",
			expectError: true,
			errMsg:      "phone number must start with 08 or 62",
		},
		{
			name:        "Too short (< 10 digits after normalization)",
			input:       "0812345",
			expectError: true,
			errMsg:      "invalid phone number length",
		},
		{
			name:        "Too long (> 13 digits after normalization)",
			input:       "081234567890123",
			expectError: true,
			errMsg:      "invalid phone number length",
		},
		{
			name:        "Empty string",
			input:       "",
			expectError: true,
			errMsg:      "phone number must start with 08 or 62",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := phonenumber.Normalize(tt.input)
			if tt.expectError {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestNormalize_Concurrency(t *testing.T) {
	const goroutines = 100
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			input := fmt.Sprintf("+62 812-3456-%04d", id%10000)
			expected := fmt.Sprintf("628123456%04d", id%10000)
			res, err := phonenumber.Normalize(input)
			assert.NoError(t, err)
			assert.Equal(t, expected, res)
		}(i)
	}

	wg.Wait()
}
