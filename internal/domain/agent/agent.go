package agent

import (
	"errors"
	"fmt"
	"time"

	"github.com/Ecom-micro-template/service-agent/internal/domain/shared"
)

// Domain errors for Agent aggregate
var (
	ErrAgentNotFound  = errors.New("agent not found")
	ErrInvalidAgent   = errors.New("invalid agent data")
	ErrAgentSuspended = errors.New("agent is suspended")
	ErrAgentInactive  = errors.New("agent is inactive")
	ErrEmailExists    = errors.New("email already registered")
)

// Agent is the aggregate root for sales agents.
type Agent struct {
	id             uint
	code           string
	name           string
	email          string
	phone          string
	commissionRate shared.CommissionRate
	tier           shared.AgentTier
	status         shared.AgentStatus
	totalEarned    float64
	teamID         *uint
	createdAt      time.Time
	updatedAt      time.Time

	// Domain events
	events []Event
}

// AgentParams contains parameters for creating an Agent.
type AgentParams struct {
	ID             uint
	Code           string
	Name           string
	Email          string
	Phone          string
	CommissionRate float64
	Tier           string
	Status         string
	TeamID         *uint
}

// NewAgent creates a new Agent aggregate.
func NewAgent(params AgentParams) (*Agent, error) {
	if params.Name == "" {
		return nil, errors.New("name is required")
	}
	if params.Email == "" {
		return nil, errors.New("email is required")
	}

	code := params.Code
	if code == "" {
		code = fmt.Sprintf("AGT%04d", params.ID)
	}

	rate := shared.DefaultCommissionRate()
	if params.CommissionRate > 0 {
		r, err := shared.NewCommissionRate(params.CommissionRate)
		if err == nil {
			rate = r
		}
	}

	tier := shared.TierBronze
	if params.Tier != "" {
		t, err := shared.ParseAgentTier(params.Tier)
		if err == nil {
			tier = t
		}
	}

	status := shared.AgentStatusActive
	if params.Status != "" {
		s, err := shared.ParseAgentStatus(params.Status)
		if err == nil {
			status = s
		}
	}

	now := time.Now()
	agent := &Agent{
		id:             params.ID,
		code:           code,
		name:           params.Name,
		email:          params.Email,
		phone:          params.Phone,
		commissionRate: rate,
		tier:           tier,
		status:         status,
		totalEarned:    0,
		teamID:         params.TeamID,
		createdAt:      now,
		updatedAt:      now,
		events:         make([]Event, 0),
	}

	agent.addEvent(NewAgentCreatedEvent(params.ID, code, params.Name))

	return agent, nil
}

// Getters
func (a *Agent) ID() uint                              { return a.id }
func (a *Agent) Code() string                          { return a.code }
func (a *Agent) Name() string                          { return a.name }
func (a *Agent) Email() string                         { return a.email }
func (a *Agent) Phone() string                         { return a.phone }
func (a *Agent) CommissionRate() shared.CommissionRate { return a.commissionRate }
func (a *Agent) Tier() shared.AgentTier                { return a.tier }
func (a *Agent) Status() shared.AgentStatus            { return a.status }
func (a *Agent) TotalEarned() float64                  { return a.totalEarned }
func (a *Agent) TeamID() *uint                         { return a.teamID }
func (a *Agent) CreatedAt() time.Time                  { return a.createdAt }
func (a *Agent) UpdatedAt() time.Time                  { return a.updatedAt }

// EffectiveCommissionRate returns rate including tier bonus.
func (a *Agent) EffectiveCommissionRate() shared.CommissionRate {
	bonus := a.tier.BonusPercentage()
	return a.commissionRate.AddPercentage(bonus)
}

// --- Behavior Methods ---

// UpdateProfile updates the agent's profile.
func (a *Agent) UpdateProfile(name, email, phone string) error {
	if name != "" {
		a.name = name
	}
	if email != "" {
		a.email = email
	}
	a.phone = phone
	a.updatedAt = time.Now()
	return nil
}

// SetCommissionRate sets the commission rate.
func (a *Agent) SetCommissionRate(rate float64) error {
	r, err := shared.NewCommissionRate(rate)
	if err != nil {
		return err
	}
	a.commissionRate = r
	a.updatedAt = time.Now()
	return nil
}

// PromoteTier promotes the agent to the next tier.
func (a *Agent) PromoteTier() error {
	nextTier := a.tier.NextTier()
	if nextTier == a.tier {
		return errors.New("already at highest tier")
	}
	a.tier = nextTier
	a.updatedAt = time.Now()
	a.addEvent(NewAgentPromotedEvent(a.id, string(nextTier)))
	return nil
}

// SetTier sets the agent tier directly.
func (a *Agent) SetTier(tierStr string) error {
	tier, err := shared.ParseAgentTier(tierStr)
	if err != nil {
		return err
	}
	a.tier = tier
	a.updatedAt = time.Now()
	return nil
}

// Activate activates the agent.
func (a *Agent) Activate() error {
	if !a.status.CanBeActivated() {
		return ErrAgentInactive
	}
	a.status = shared.AgentStatusActive
	a.updatedAt = time.Now()
	a.addEvent(NewAgentStatusChangedEvent(a.id, string(a.status)))
	return nil
}

// Suspend suspends the agent.
func (a *Agent) Suspend(reason string) error {
	if a.status == shared.AgentStatusSuspended {
		return nil
	}
	a.status = shared.AgentStatusSuspended
	a.updatedAt = time.Now()
	a.addEvent(NewAgentStatusChangedEvent(a.id, string(a.status)))
	return nil
}

// Deactivate deactivates the agent.
func (a *Agent) Deactivate() error {
	a.status = shared.AgentStatusInactive
	a.updatedAt = time.Now()
	a.addEvent(NewAgentStatusChangedEvent(a.id, string(a.status)))
	return nil
}

// AssignToTeam assigns the agent to a team.
func (a *Agent) AssignToTeam(teamID uint) {
	a.teamID = &teamID
	a.updatedAt = time.Now()
}

// RemoveFromTeam removes the agent from their team.
func (a *Agent) RemoveFromTeam() {
	a.teamID = nil
	a.updatedAt = time.Now()
}

// RecordEarnings records earnings from a commission.
func (a *Agent) RecordEarnings(amount float64) {
	a.totalEarned += amount
	a.updatedAt = time.Now()
}

// CanEarnCommission returns true if agent can earn commissions.
func (a *Agent) CanEarnCommission() bool {
	return a.status.CanEarnCommission()
}

// CanReceivePayout returns true if agent can receive payouts.
func (a *Agent) CanReceivePayout() bool {
	return a.status.CanReceivePayout()
}

// IsActive returns true if agent is active.
func (a *Agent) IsActive() bool {
	return a.status == shared.AgentStatusActive
}

// Events returns and clears the collected domain events.
func (a *Agent) Events() []Event {
	events := a.events
	a.events = make([]Event, 0)
	return events
}

func (a *Agent) addEvent(event Event) {
	a.events = append(a.events, event)
}
