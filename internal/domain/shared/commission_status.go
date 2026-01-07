package shared

import (
	"errors"
	"fmt"
)

// CommissionStatus represents the status of a commission.
type CommissionStatus string

// Commission status constants
const (
	CommissionPending   CommissionStatus = "pending"
	CommissionApproved  CommissionStatus = "approved"
	CommissionPaid      CommissionStatus = "paid"
	CommissionCancelled CommissionStatus = "cancelled"
)

// validCommissionTransitions defines allowed state transitions.
var validCommissionTransitions = map[CommissionStatus][]CommissionStatus{
	CommissionPending:   {CommissionApproved, CommissionCancelled},
	CommissionApproved:  {CommissionPaid, CommissionCancelled},
	CommissionPaid:      {}, // Terminal
	CommissionCancelled: {}, // Terminal
}

// ErrInvalidCommissionStatus is returned for invalid status values.
var ErrInvalidCommissionStatus = errors.New("invalid commission status")

// ErrInvalidCommissionTransition is returned for invalid transitions.
var ErrInvalidCommissionTransition = errors.New("invalid commission status transition")

// AllCommissionStatuses returns all valid statuses.
func AllCommissionStatuses() []CommissionStatus {
	return []CommissionStatus{CommissionPending, CommissionApproved, CommissionPaid, CommissionCancelled}
}

// IsValid returns true if the status is valid.
func (s CommissionStatus) IsValid() bool {
	switch s {
	case CommissionPending, CommissionApproved, CommissionPaid, CommissionCancelled:
		return true
	default:
		return false
	}
}

// String returns the string representation.
func (s CommissionStatus) String() string {
	return string(s)
}

// Label returns a human-readable label.
func (s CommissionStatus) Label() string {
	switch s {
	case CommissionPending:
		return "Pending"
	case CommissionApproved:
		return "Approved"
	case CommissionPaid:
		return "Paid"
	case CommissionCancelled:
		return "Cancelled"
	default:
		return "Unknown"
	}
}

// CanTransitionTo returns true if the status can transition to target.
func (s CommissionStatus) CanTransitionTo(target CommissionStatus) bool {
	allowed, exists := validCommissionTransitions[s]
	if !exists {
		return false
	}
	for _, status := range allowed {
		if status == target {
			return true
		}
	}
	return false
}

// TransitionTo attempts to transition to the target status.
func (s CommissionStatus) TransitionTo(target CommissionStatus) (CommissionStatus, error) {
	if !s.CanTransitionTo(target) {
		return s, fmt.Errorf("%w: cannot transition from %s to %s", ErrInvalidCommissionTransition, s, target)
	}
	return target, nil
}

// IsPending returns true if status is pending.
func (s CommissionStatus) IsPending() bool {
	return s == CommissionPending
}

// IsApproved returns true if status is approved.
func (s CommissionStatus) IsApproved() bool {
	return s == CommissionApproved
}

// IsPaid returns true if status is paid.
func (s CommissionStatus) IsPaid() bool {
	return s == CommissionPaid
}

// IsPayable returns true if commission can be included in a payout.
func (s CommissionStatus) IsPayable() bool {
	return s == CommissionApproved
}

// IsTerminal returns true if status is terminal.
func (s CommissionStatus) IsTerminal() bool {
	return s == CommissionPaid || s == CommissionCancelled
}

// ParseCommissionStatus parses a string into a CommissionStatus.
func ParseCommissionStatus(str string) (CommissionStatus, error) {
	s := CommissionStatus(str)
	if !s.IsValid() {
		return "", fmt.Errorf("%w: %s", ErrInvalidCommissionStatus, str)
	}
	return s, nil
}
