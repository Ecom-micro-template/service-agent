// Package models contains GORM persistence models for the agent service.
//
// Deprecated: This package is being migrated to DDD architecture.
// For new development, use:
//   - Domain models: github.com/Ecom-micro-template/service-agent/internal/domain/payout
//   - Persistence: github.com/Ecom-micro-template/service-agent/internal/infrastructure/persistence
//
// Existing code can continue using this package during the transition period.
package models

import (
	"time"
)

type Payout struct {
	ID            uint       `gorm:"primaryKey" json:"id"`
	AgentID       uint       `gorm:"not null;index" json:"agent_id"`
	Amount        float64    `gorm:"type:decimal(10,2);not null" json:"amount"`
	Period        string     `gorm:"size:20;not null" json:"period"`  // Format: YYYY-MM
	CommissionIDs string     `gorm:"type:text" json:"commission_ids"` // JSON array of commission IDs
	Status        string     `gorm:"size:20;default:'pending'" json:"status"`
	PaidAt        *time.Time `json:"paid_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`

	// Relations
	Agent Agent `gorm:"foreignKey:AgentID" json:"agent,omitempty"`
}

func (Payout) TableName() string {
	return "payouts"
}
