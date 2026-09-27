package entity_test

import (
	"testing"

	"github.com/mrbayss/golang-simple-ecommerce/internal/entity"
	"github.com/stretchr/testify/assert"
)

func TestMoney_IDR(t *testing.T) {
	tests := []struct {
		name     string
		money    entity.Money
		expected string
	}{
		{
			name:     "Zero amount",
			money:    entity.Money(0),
			expected: "Rp0",
		},
		{
			name:     "Standard amount",
			money:    entity.Money(50000),
			expected: "Rp50000",
		},
		{
			name:     "Large amount",
			money:    entity.Money(1500000000),
			expected: "Rp1500000000",
		},
		{
			name:     "Negative amount",
			money:    entity.Money(-10000),
			expected: "Rp-10000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.money.IDR())
		})
	}
}
