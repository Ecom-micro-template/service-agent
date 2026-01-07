package shared

import (
	"errors"
	"fmt"
)

// CommissionRate represents a commission rate percentage.
// Rate is stored as a decimal (0-100).
type CommissionRate struct {
	value float64
}

// ErrInvalidCommissionRate is returned for invalid rates.
var ErrInvalidCommissionRate = errors.New("commission rate must be between 0 and 100")

// NewCommissionRate creates a new CommissionRate with validation.
func NewCommissionRate(rate float64) (CommissionRate, error) {
	if rate < 0 || rate > 100 {
		return CommissionRate{}, ErrInvalidCommissionRate
	}
	return CommissionRate{value: rate}, nil
}

// MustCommissionRate creates a CommissionRate, panicking on error.
func MustCommissionRate(rate float64) CommissionRate {
	r, err := NewCommissionRate(rate)
	if err != nil {
		panic(err)
	}
	return r
}

// DefaultCommissionRate returns the default commission rate (10%).
func DefaultCommissionRate() CommissionRate {
	return CommissionRate{value: 10.0}
}

// Value returns the rate as a percentage (0-100).
func (r CommissionRate) Value() float64 {
	return r.value
}

// Percentage returns the rate as a decimal (0-1).
func (r CommissionRate) Percentage() float64 {
	return r.value / 100.0
}

// String returns the string representation.
func (r CommissionRate) String() string {
	return fmt.Sprintf("%.2f%%", r.value)
}

// IsZero returns true if the rate is zero.
func (r CommissionRate) IsZero() bool {
	return r.value == 0
}

// CalculateCommission calculates commission from an amount.
func (r CommissionRate) CalculateCommission(amount float64) float64 {
	return amount * r.Percentage()
}

// Add adds another rate to this rate (capped at 100).
func (r CommissionRate) Add(other CommissionRate) CommissionRate {
	newValue := r.value + other.value
	if newValue > 100 {
		newValue = 100
	}
	return CommissionRate{value: newValue}
}

// AddPercentage adds a percentage bonus (e.g., 2% from tier).
func (r CommissionRate) AddPercentage(bonus float64) CommissionRate {
	newValue := r.value + bonus*100
	if newValue > 100 {
		newValue = 100
	}
	return CommissionRate{value: newValue}
}

// Equals compares two rates.
func (r CommissionRate) Equals(other CommissionRate) bool {
	return r.value == other.value
}

// IsHigherThan returns true if this rate is higher than other.
func (r CommissionRate) IsHigherThan(other CommissionRate) bool {
	return r.value > other.value
}
