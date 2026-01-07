package commission

import (
	"time"
)

// Event is the base interface for all commission domain events.
type Event interface {
	EventType() string
	OccurredAt() time.Time
}

// baseEvent contains common event fields.
type baseEvent struct {
	occurredAt time.Time
}

func (e baseEvent) OccurredAt() time.Time { return e.occurredAt }

// CommissionCreatedEvent is raised when a new commission is created.
type CommissionCreatedEvent struct {
	baseEvent
	CommissionID uint
	AgentID      uint
	OrderID      string
	Amount       float64
}

func (e CommissionCreatedEvent) EventType() string { return "commission.created" }

// NewCommissionCreatedEvent creates a new CommissionCreatedEvent.
func NewCommissionCreatedEvent(commissionID, agentID uint, orderID string, amount float64) CommissionCreatedEvent {
	return CommissionCreatedEvent{
		baseEvent:    baseEvent{occurredAt: time.Now()},
		CommissionID: commissionID,
		AgentID:      agentID,
		OrderID:      orderID,
		Amount:       amount,
	}
}

// CommissionApprovedEvent is raised when a commission is approved.
type CommissionApprovedEvent struct {
	baseEvent
	CommissionID uint
	AgentID      uint
	Amount       float64
}

func (e CommissionApprovedEvent) EventType() string { return "commission.approved" }

// NewCommissionApprovedEvent creates a new CommissionApprovedEvent.
func NewCommissionApprovedEvent(commissionID, agentID uint, amount float64) CommissionApprovedEvent {
	return CommissionApprovedEvent{
		baseEvent:    baseEvent{occurredAt: time.Now()},
		CommissionID: commissionID,
		AgentID:      agentID,
		Amount:       amount,
	}
}

// CommissionPaidEvent is raised when a commission is paid.
type CommissionPaidEvent struct {
	baseEvent
	CommissionID uint
	AgentID      uint
	Amount       float64
}

func (e CommissionPaidEvent) EventType() string { return "commission.paid" }

// NewCommissionPaidEvent creates a new CommissionPaidEvent.
func NewCommissionPaidEvent(commissionID, agentID uint, amount float64) CommissionPaidEvent {
	return CommissionPaidEvent{
		baseEvent:    baseEvent{occurredAt: time.Now()},
		CommissionID: commissionID,
		AgentID:      agentID,
		Amount:       amount,
	}
}
