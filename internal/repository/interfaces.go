package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/Ecom-micro-template/service-agent/internal/domain"
)

// =============================================================================
// INTERFACE SEGREGATION PRINCIPLE (ISP)
// =============================================================================
//
// This file defines segregated interfaces following ISP. Each interface
// represents a specific capability, allowing consumers to depend only on
// the methods they actually use.
//
// Benefits:
// - Handlers depend only on methods they need
// - Easier to mock in tests
// - Clear separation of read/write operations
// - Supports CQRS patterns
//
// =============================================================================

// =============================================================================
// AGENT REPOSITORY INTERFACES
// =============================================================================

// AgentReader provides read-only access to agents
// Use this for handlers that only need to query agents
type AgentReader interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Agent, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.Agent, error)
	GetByCode(ctx context.Context, code string) (*domain.Agent, error)
	List(ctx context.Context, filter AgentFilter) ([]domain.Agent, int64, error)
	GetStats(ctx context.Context, agentID uuid.UUID) (*AgentStats, error)
}

// AgentWriter provides write access to agents
// Use this for handlers that create or modify agents
type AgentWriter interface {
	Create(ctx context.Context, agent *domain.Agent) error
	Update(ctx context.Context, agent *domain.Agent) error
	Delete(ctx context.Context, id uuid.UUID) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
	UpdateTier(ctx context.Context, id uuid.UUID, tier string) error
}

// AgentRepository is the composed interface
type AgentRepository interface {
	AgentReader
	AgentWriter
}

// AgentFilter represents filters for listing agents
type AgentFilter struct {
	Status   string
	Tier     string
	ParentID *uuid.UUID
	Search   string
	Page     int
	Limit    int
}

// AgentStats represents agent statistics
type AgentStats struct {
	TotalOrders       int64   `json:"total_orders"`
	TotalCommission   float64 `json:"total_commission"`
	PendingCommission float64 `json:"pending_commission"`
	PaidCommission    float64 `json:"paid_commission"`
	TotalCustomers    int64   `json:"total_customers"`
}

// =============================================================================
// COMMISSION REPOSITORY INTERFACES
// =============================================================================

// CommissionReader provides read-only access to commissions
type CommissionReader interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Commission, error)
	GetByAgentID(ctx context.Context, agentID uuid.UUID, page, limit int) ([]domain.Commission, int64, error)
	GetByOrderID(ctx context.Context, orderID uuid.UUID) ([]domain.Commission, error)
	GetPending(ctx context.Context, page, limit int) ([]domain.Commission, int64, error)
}

// CommissionWriter provides write access to commissions
type CommissionWriter interface {
	Create(ctx context.Context, commission *domain.Commission) error
	Update(ctx context.Context, commission *domain.Commission) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
	Approve(ctx context.Context, id uuid.UUID, approvedBy uuid.UUID) error
}

// CommissionCalculator calculates commission rates
type CommissionCalculator interface {
	CalculateCommission(ctx context.Context, agentID uuid.UUID, orderAmount float64, categoryID *uuid.UUID) (float64, error)
}

// CommissionRepository is the composed interface
type CommissionRepository interface {
	CommissionReader
	CommissionWriter
	CommissionCalculator
}

// =============================================================================
// CATEGORY COMMISSION REPOSITORY INTERFACES
// =============================================================================

// CategoryCommissionReader provides read-only access to category commissions
type CategoryCommissionReader interface {
	GetByAgentID(ctx context.Context, agentID uuid.UUID) ([]domain.AgentCategoryCommission, error)
	GetByAgentAndCategory(ctx context.Context, agentID, categoryID uuid.UUID) (*domain.AgentCategoryCommission, error)
}

// CategoryCommissionWriter provides write access to category commissions
type CategoryCommissionWriter interface {
	Upsert(ctx context.Context, commission *domain.AgentCategoryCommission) error
	Delete(ctx context.Context, agentID, categoryID uuid.UUID) error
	BulkUpdate(ctx context.Context, agentID uuid.UUID, commissions []domain.AgentCategoryCommission) error
}

// CategoryCommissionRepository is the composed interface
type CategoryCommissionRepository interface {
	CategoryCommissionReader
	CategoryCommissionWriter
}

// =============================================================================
// PAYOUT REPOSITORY INTERFACES
// =============================================================================

// PayoutReader provides read-only access to payouts
type PayoutReader interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Payout, error)
	GetByAgentID(ctx context.Context, agentID uuid.UUID, page, limit int) ([]domain.Payout, int64, error)
	GetPending(ctx context.Context, page, limit int) ([]domain.Payout, int64, error)
}

// PayoutWriter provides write access to payouts
type PayoutWriter interface {
	Create(ctx context.Context, payout *domain.Payout) error
	Update(ctx context.Context, payout *domain.Payout) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
	MarkAsPaid(ctx context.Context, id uuid.UUID, paidBy uuid.UUID, transactionRef string) error
}

// PayoutRepository is the composed interface
type PayoutRepository interface {
	PayoutReader
	PayoutWriter
}

// =============================================================================
// TEAM REPOSITORY INTERFACES
// =============================================================================

// TeamReader provides read-only access to teams
type TeamReader interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Team, error)
	GetByAgentID(ctx context.Context, agentID uuid.UUID) (*domain.Team, error)
	GetMembers(ctx context.Context, teamID uuid.UUID) ([]domain.Agent, error)
}

// TeamWriter provides write access to teams
type TeamWriter interface {
	Create(ctx context.Context, team *domain.Team) error
	Update(ctx context.Context, team *domain.Team) error
	AddMember(ctx context.Context, teamID, agentID uuid.UUID) error
	RemoveMember(ctx context.Context, teamID, agentID uuid.UUID) error
}

// TeamRepository is the composed interface
type TeamRepository interface {
	TeamReader
	TeamWriter
}
