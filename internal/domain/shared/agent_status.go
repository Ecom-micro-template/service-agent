// Package shared provides shared value objects for the agent domain.
package shared

import (
	"errors"
	"fmt"
)

// AgentStatus represents the status of an agent.
type AgentStatus string

// Agent status constants
const (
	AgentStatusActive    AgentStatus = "active"
	AgentStatusInactive  AgentStatus = "inactive"
	AgentStatusSuspended AgentStatus = "suspended"
	AgentStatusPending   AgentStatus = "pending"
)

// ErrInvalidAgentStatus is returned for invalid status values.
var ErrInvalidAgentStatus = errors.New("invalid agent status")

// AllAgentStatuses returns all valid statuses.
func AllAgentStatuses() []AgentStatus {
	return []AgentStatus{AgentStatusActive, AgentStatusInactive, AgentStatusSuspended, AgentStatusPending}
}

// IsValid returns true if the status is valid.
func (s AgentStatus) IsValid() bool {
	switch s {
	case AgentStatusActive, AgentStatusInactive, AgentStatusSuspended, AgentStatusPending:
		return true
	default:
		return false
	}
}

// String returns the string representation.
func (s AgentStatus) String() string {
	return string(s)
}

// Label returns a human-readable label.
func (s AgentStatus) Label() string {
	switch s {
	case AgentStatusActive:
		return "Active"
	case AgentStatusInactive:
		return "Inactive"
	case AgentStatusSuspended:
		return "Suspended"
	case AgentStatusPending:
		return "Pending"
	default:
		return "Unknown"
	}
}

// CanEarnCommission returns true if agent can earn commissions.
func (s AgentStatus) CanEarnCommission() bool {
	return s == AgentStatusActive
}

// CanReceivePayout returns true if agent can receive payouts.
func (s AgentStatus) CanReceivePayout() bool {
	return s == AgentStatusActive
}

// CanBeActivated returns true if status can transition to active.
func (s AgentStatus) CanBeActivated() bool {
	return s == AgentStatusInactive || s == AgentStatusPending
}

// ParseAgentStatus parses a string into an AgentStatus.
func ParseAgentStatus(s string) (AgentStatus, error) {
	status := AgentStatus(s)
	if !status.IsValid() {
		return "", fmt.Errorf("%w: %s", ErrInvalidAgentStatus, s)
	}
	return status, nil
}
