package shared

import (
	"errors"
	"fmt"
)

// PayoutStatus represents the status of a payout.
type PayoutStatus string

// Payout status constants
const (
	PayoutPending    PayoutStatus = "pending"
	PayoutProcessing PayoutStatus = "processing"
	PayoutCompleted  PayoutStatus = "completed"
	PayoutFailed     PayoutStatus = "failed"
	PayoutCancelled  PayoutStatus = "cancelled"
)

// validPayoutTransitions defines allowed state transitions.
var validPayoutTransitions = map[PayoutStatus][]PayoutStatus{
	PayoutPending:    {PayoutProcessing, PayoutCancelled},
	PayoutProcessing: {PayoutCompleted, PayoutFailed},
	PayoutCompleted:  {},              // Terminal
	PayoutFailed:     {PayoutPending}, // Can retry
	PayoutCancelled:  {},              // Terminal
}

// ErrInvalidPayoutStatus is returned for invalid status values.
var ErrInvalidPayoutStatus = errors.New("invalid payout status")

// ErrInvalidPayoutTransition is returned for invalid transitions.
var ErrInvalidPayoutTransition = errors.New("invalid payout status transition")

// AllPayoutStatuses returns all valid statuses.
func AllPayoutStatuses() []PayoutStatus {
	return []PayoutStatus{PayoutPending, PayoutProcessing, PayoutCompleted, PayoutFailed, PayoutCancelled}
}

// IsValid returns true if the status is valid.
func (s PayoutStatus) IsValid() bool {
	switch s {
	case PayoutPending, PayoutProcessing, PayoutCompleted, PayoutFailed, PayoutCancelled:
		return true
	default:
		return false
	}
}

// String returns the string representation.
func (s PayoutStatus) String() string {
	return string(s)
}

// Label returns a human-readable label.
func (s PayoutStatus) Label() string {
	switch s {
	case PayoutPending:
		return "Pending"
	case PayoutProcessing:
		return "Processing"
	case PayoutCompleted:
		return "Completed"
	case PayoutFailed:
		return "Failed"
	case PayoutCancelled:
		return "Cancelled"
	default:
		return "Unknown"
	}
}

// CanTransitionTo returns true if the status can transition to target.
func (s PayoutStatus) CanTransitionTo(target PayoutStatus) bool {
	allowed, exists := validPayoutTransitions[s]
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
func (s PayoutStatus) TransitionTo(target PayoutStatus) (PayoutStatus, error) {
	if !s.CanTransitionTo(target) {
		return s, fmt.Errorf("%w: cannot transition from %s to %s", ErrInvalidPayoutTransition, s, target)
	}
	return target, nil
}

// IsPending returns true if status is pending.
func (s PayoutStatus) IsPending() bool {
	return s == PayoutPending
}

// IsProcessing returns true if status is processing.
func (s PayoutStatus) IsProcessing() bool {
	return s == PayoutProcessing
}

// IsCompleted returns true if status is completed.
func (s PayoutStatus) IsCompleted() bool {
	return s == PayoutCompleted
}

// IsFailed returns true if status is failed.
func (s PayoutStatus) IsFailed() bool {
	return s == PayoutFailed
}

// IsTerminal returns true if status is terminal.
func (s PayoutStatus) IsTerminal() bool {
	return s == PayoutCompleted || s == PayoutCancelled
}

// CanRetry returns true if payout can be retried.
func (s PayoutStatus) CanRetry() bool {
	return s == PayoutFailed
}

// ParsePayoutStatus parses a string into a PayoutStatus.
func ParsePayoutStatus(str string) (PayoutStatus, error) {
	s := PayoutStatus(str)
	if !s.IsValid() {
		return "", fmt.Errorf("%w: %s", ErrInvalidPayoutStatus, str)
	}
	return s, nil
}
