package agent

import (
	"time"
)

// Event is the base interface for all agent domain events.
type Event interface {
	EventType() string
	OccurredAt() time.Time
	AggregateID() uint
}

// baseEvent contains common event fields.
type baseEvent struct {
	occurredAt  time.Time
	aggregateID uint
}

func (e baseEvent) OccurredAt() time.Time { return e.occurredAt }
func (e baseEvent) AggregateID() uint     { return e.aggregateID }

// AgentCreatedEvent is raised when a new agent is created.
type AgentCreatedEvent struct {
	baseEvent
	Code string
	Name string
}

func (e AgentCreatedEvent) EventType() string { return "agent.created" }

// NewAgentCreatedEvent creates a new AgentCreatedEvent.
func NewAgentCreatedEvent(agentID uint, code, name string) AgentCreatedEvent {
	return AgentCreatedEvent{
		baseEvent: baseEvent{occurredAt: time.Now(), aggregateID: agentID},
		Code:      code,
		Name:      name,
	}
}

// AgentStatusChangedEvent is raised when agent status changes.
type AgentStatusChangedEvent struct {
	baseEvent
	NewStatus string
}

func (e AgentStatusChangedEvent) EventType() string { return "agent.status_changed" }

// NewAgentStatusChangedEvent creates a new AgentStatusChangedEvent.
func NewAgentStatusChangedEvent(agentID uint, newStatus string) AgentStatusChangedEvent {
	return AgentStatusChangedEvent{
		baseEvent: baseEvent{occurredAt: time.Now(), aggregateID: agentID},
		NewStatus: newStatus,
	}
}

// AgentPromotedEvent is raised when agent is promoted to a new tier.
type AgentPromotedEvent struct {
	baseEvent
	NewTier string
}

func (e AgentPromotedEvent) EventType() string { return "agent.promoted" }

// NewAgentPromotedEvent creates a new AgentPromotedEvent.
func NewAgentPromotedEvent(agentID uint, newTier string) AgentPromotedEvent {
	return AgentPromotedEvent{
		baseEvent: baseEvent{occurredAt: time.Now(), aggregateID: agentID},
		NewTier:   newTier,
	}
}
