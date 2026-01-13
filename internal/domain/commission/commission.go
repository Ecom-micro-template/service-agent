package commission

import (
	"errors"
	"time"

	"github.com/Ecom-micro-template/service-agent/internal/domain/shared"
)

// Domain errors for Commission entity
var (
	ErrCommissionNotFound = errors.New("commission not found")
	ErrInvalidCommission  = errors.New("invalid commission data")
	ErrAlreadyPaid        = errors.New("commission already paid")
	ErrNotApproved        = errors.New("commission must be approved before payment")
)

// Commission represents an agent's commission for an order.
type Commission struct {
	id         uint
	agentID    uint
	orderID    string
	orderTotal float64
	rate       shared.CommissionRate
	amount     float64
	status     shared.CommissionStatus
	createdAt  time.Time
	updatedAt  time.Time

	// Domain events
	events []Event
}

// CommissionParams contains parameters for creating a Commission.
type CommissionParams struct {
	ID         uint
	AgentID    uint
	OrderID    string
	OrderTotal float64
	Rate       float64
	Amount     float64
}

// NewCommission creates a new Commission entity.
func NewCommission(params CommissionParams) (*Commission, error) {
	if params.AgentID == 0 {
		return nil, errors.New("agent ID is required")
	}
	if params.OrderID == "" {
		return nil, errors.New("order ID is required")
	}
	if params.OrderTotal <= 0 {
		return nil, errors.New("order total must be positive")
	}

	rate, err := shared.NewCommissionRate(params.Rate)
	if err != nil {
		return nil, err
	}

	// Calculate amount if not provided
	amount := params.Amount
	if amount <= 0 {
		amount = rate.CalculateCommission(params.OrderTotal)
	}

	now := time.Now()
	commission := &Commission{
		id:         params.ID,
		agentID:    params.AgentID,
		orderID:    params.OrderID,
		orderTotal: params.OrderTotal,
		rate:       rate,
		amount:     amount,
		status:     shared.CommissionPending,
		createdAt:  now,
		updatedAt:  now,
		events:     make([]Event, 0),
	}

	commission.addEvent(NewCommissionCreatedEvent(params.ID, params.AgentID, params.OrderID, amount))

	return commission, nil
}

// Getters
func (c *Commission) ID() uint                        { return c.id }
func (c *Commission) AgentID() uint                   { return c.agentID }
func (c *Commission) OrderID() string                 { return c.orderID }
func (c *Commission) OrderTotal() float64             { return c.orderTotal }
func (c *Commission) Rate() shared.CommissionRate     { return c.rate }
func (c *Commission) Amount() float64                 { return c.amount }
func (c *Commission) Status() shared.CommissionStatus { return c.status }
func (c *Commission) CreatedAt() time.Time            { return c.createdAt }
func (c *Commission) UpdatedAt() time.Time            { return c.updatedAt }

// --- Behavior Methods ---

// Approve approves the commission for payment.
func (c *Commission) Approve() error {
	if !c.status.CanTransitionTo(shared.CommissionApproved) {
		return ErrInvalidCommission
	}
	c.status = shared.CommissionApproved
	c.updatedAt = time.Now()
	c.addEvent(NewCommissionApprovedEvent(c.id, c.agentID, c.amount))
	return nil
}

// MarkAsPaid marks the commission as paid.
func (c *Commission) MarkAsPaid() error {
	if !c.status.CanTransitionTo(shared.CommissionPaid) {
		return ErrNotApproved
	}
	c.status = shared.CommissionPaid
	c.updatedAt = time.Now()
	c.addEvent(NewCommissionPaidEvent(c.id, c.agentID, c.amount))
	return nil
}

// Cancel cancels the commission.
func (c *Commission) Cancel(reason string) error {
	if c.status.IsTerminal() {
		return ErrAlreadyPaid
	}
	c.status = shared.CommissionCancelled
	c.updatedAt = time.Now()
	return nil
}

// IsPending returns true if commission is pending.
func (c *Commission) IsPending() bool {
	return c.status.IsPending()
}

// IsApproved returns true if commission is approved.
func (c *Commission) IsApproved() bool {
	return c.status.IsApproved()
}

// IsPaid returns true if commission is paid.
func (c *Commission) IsPaid() bool {
	return c.status.IsPaid()
}

// IsPayable returns true if commission can be included in a payout.
func (c *Commission) IsPayable() bool {
	return c.status.IsPayable()
}

// Events returns and clears the collected domain events.
func (c *Commission) Events() []Event {
	events := c.events
	c.events = make([]Event, 0)
	return events
}

func (c *Commission) addEvent(event Event) {
	c.events = append(c.events, event)
}
