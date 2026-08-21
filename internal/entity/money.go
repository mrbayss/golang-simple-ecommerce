package entity

import "fmt"

// Money is the canonical type for all monetary values.
// Stored as bigint in Postgres, unit: whole Rupiah (no cents).
type Money int64

// IDR renders the value for display purposes.
func (m Money) IDR() string {
	return fmt.Sprintf("Rp%d", int64(m))
}
