package payout

import (
	"errors"
	"time"

	"github.com/niaga-platform/service-agent/internal/domain/shared"
)

// Domain errors for Payout aggregate
var (
	ErrPayoutNotFound  = errors.New("payout not found")
	ErrInvalidPayout   = errors.New("invalid payout data")
	ErrPayoutCompleted = errors.New("payout already completed")
	ErrNoCommissions   = errors.New("no commissions to payout")
)

// Payout is the aggregate root for agent payouts.
type Payout struct {
	id        uint
	agentID   uint
	amount    float64
	period    string // Format: YYYY-MM
	items     []PayoutItem
	status    shared.PayoutStatus
	paidAt    *time.Time
	createdAt time.Time
	updatedAt time.Time
}

// PayoutParams contains parameters for creating a Payout.
type PayoutParams struct {
	ID      uint
	AgentID uint
	Period  string
	Items   []PayoutItem
}

// NewPayout creates a new Payout aggregate.
func NewPayout(params PayoutParams) (*Payout, error) {
	if params.AgentID == 0 {
		return nil, errors.New("agent ID is required")
	}
	if params.Period == "" {
		return nil, errors.New("period is required")
	}
	if len(params.Items) == 0 {
		return nil, ErrNoCommissions
	}

	// Calculate total amount
	var amount float64
	for _, item := range params.Items {
		amount += item.Amount()
	}

	now := time.Now()
	return &Payout{
		id:        params.ID,
		agentID:   params.AgentID,
		amount:    amount,
		period:    params.Period,
		items:     params.Items,
		status:    shared.PayoutPending,
		createdAt: now,
		updatedAt: now,
	}, nil
}

// Getters
func (p *Payout) ID() uint                    { return p.id }
func (p *Payout) AgentID() uint               { return p.agentID }
func (p *Payout) Amount() float64             { return p.amount }
func (p *Payout) Period() string              { return p.period }
func (p *Payout) Items() []PayoutItem         { return p.items }
func (p *Payout) Status() shared.PayoutStatus { return p.status }
func (p *Payout) PaidAt() *time.Time          { return p.paidAt }
func (p *Payout) CreatedAt() time.Time        { return p.createdAt }
func (p *Payout) UpdatedAt() time.Time        { return p.updatedAt }

// CommissionIDs returns all commission IDs in this payout.
func (p *Payout) CommissionIDs() []uint {
	ids := make([]uint, len(p.items))
	for i, item := range p.items {
		ids[i] = item.CommissionID()
	}
	return ids
}

// ItemCount returns the number of commissions in this payout.
func (p *Payout) ItemCount() int {
	return len(p.items)
}

// --- Behavior Methods ---

// Process starts processing the payout.
func (p *Payout) Process() error {
	if !p.status.CanTransitionTo(shared.PayoutProcessing) {
		return ErrPayoutCompleted
	}
	p.status = shared.PayoutProcessing
	p.updatedAt = time.Now()
	return nil
}

// Complete marks the payout as completed.
func (p *Payout) Complete() error {
	if !p.status.CanTransitionTo(shared.PayoutCompleted) {
		return ErrPayoutCompleted
	}
	p.status = shared.PayoutCompleted
	now := time.Now()
	p.paidAt = &now
	p.updatedAt = now
	return nil
}

// Fail marks the payout as failed.
func (p *Payout) Fail(reason string) error {
	if !p.status.CanTransitionTo(shared.PayoutFailed) {
		return ErrPayoutCompleted
	}
	p.status = shared.PayoutFailed
	p.updatedAt = time.Now()
	return nil
}

// Retry retries a failed payout.
func (p *Payout) Retry() error {
	if !p.status.CanRetry() {
		return ErrInvalidPayout
	}
	p.status = shared.PayoutPending
	p.updatedAt = time.Now()
	return nil
}

// Cancel cancels the payout.
func (p *Payout) Cancel() error {
	if p.status.IsTerminal() {
		return ErrPayoutCompleted
	}
	p.status = shared.PayoutCancelled
	p.updatedAt = time.Now()
	return nil
}

// IsPending returns true if payout is pending.
func (p *Payout) IsPending() bool {
	return p.status.IsPending()
}

// IsCompleted returns true if payout is completed.
func (p *Payout) IsCompleted() bool {
	return p.status.IsCompleted()
}

// IsFailed returns true if payout is failed.
func (p *Payout) IsFailed() bool {
	return p.status.IsFailed()
}
