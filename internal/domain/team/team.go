package team

import (
	"errors"
	"time"
)

// Domain errors for Team entity
var (
	ErrTeamNotFound = errors.New("team not found")
	ErrInvalidTeam  = errors.New("invalid team data")
)

// Team represents a sales team.
type Team struct {
	id             uint
	code           string
	name           string
	description    string
	leaderID       *uint
	targetMonthly  float64
	commissionRate float64
	isActive       bool
	createdAt      time.Time
	updatedAt      time.Time
}

// TeamParams contains parameters for creating a Team.
type TeamParams struct {
	ID             uint
	Code           string
	Name           string
	Description    string
	LeaderID       *uint
	TargetMonthly  float64
	CommissionRate float64
	IsActive       bool
}

// NewTeam creates a new Team entity.
func NewTeam(params TeamParams) (*Team, error) {
	if params.Name == "" {
		return nil, errors.New("name is required")
	}
	if params.Code == "" {
		return nil, errors.New("code is required")
	}

	rate := params.CommissionRate
	if rate <= 0 {
		rate = 10.0 // Default 10%
	}

	now := time.Now()
	return &Team{
		id:             params.ID,
		code:           params.Code,
		name:           params.Name,
		description:    params.Description,
		leaderID:       params.LeaderID,
		targetMonthly:  params.TargetMonthly,
		commissionRate: rate,
		isActive:       params.IsActive,
		createdAt:      now,
		updatedAt:      now,
	}, nil
}

// Getters
func (t *Team) ID() uint                { return t.id }
func (t *Team) Code() string            { return t.code }
func (t *Team) Name() string            { return t.name }
func (t *Team) Description() string     { return t.description }
func (t *Team) LeaderID() *uint         { return t.leaderID }
func (t *Team) TargetMonthly() float64  { return t.targetMonthly }
func (t *Team) CommissionRate() float64 { return t.commissionRate }
func (t *Team) IsActive() bool          { return t.isActive }
func (t *Team) CreatedAt() time.Time    { return t.createdAt }
func (t *Team) UpdatedAt() time.Time    { return t.updatedAt }

// --- Behavior Methods ---

// Update updates the team details.
func (t *Team) Update(name, description string) {
	if name != "" {
		t.name = name
	}
	t.description = description
	t.updatedAt = time.Now()
}

// SetLeader sets the team leader.
func (t *Team) SetLeader(leaderID uint) {
	t.leaderID = &leaderID
	t.updatedAt = time.Now()
}

// RemoveLeader removes the team leader.
func (t *Team) RemoveLeader() {
	t.leaderID = nil
	t.updatedAt = time.Now()
}

// SetTarget sets the monthly target.
func (t *Team) SetTarget(target float64) {
	if target >= 0 {
		t.targetMonthly = target
		t.updatedAt = time.Now()
	}
}

// SetCommissionRate sets the team commission rate.
func (t *Team) SetCommissionRate(rate float64) {
	if rate >= 0 && rate <= 100 {
		t.commissionRate = rate
		t.updatedAt = time.Now()
	}
}

// Activate activates the team.
func (t *Team) Activate() {
	t.isActive = true
	t.updatedAt = time.Now()
}

// Deactivate deactivates the team.
func (t *Team) Deactivate() {
	t.isActive = false
	t.updatedAt = time.Now()
}

// HasLeader returns true if team has a leader.
func (t *Team) HasLeader() bool {
	return t.leaderID != nil
}
