// Package models contains GORM persistence models for the agent service.
//
// Deprecated: This package is being migrated to DDD architecture.
// For new development, use:
//   - Domain models: github.com/niaga-platform/service-agent/internal/domain/commission
//   - Persistence: github.com/niaga-platform/service-agent/internal/infrastructure/persistence
//
// Existing code can continue using this package during the transition period.
package models

import (
	"time"
)

type Commission struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	AgentID    uint      `gorm:"not null;index" json:"agent_id"`
	OrderID    string    `gorm:"size:100;not null;index" json:"order_id"`
	OrderTotal float64   `gorm:"type:decimal(10,2);not null" json:"order_total"`
	Rate       float64   `gorm:"type:decimal(5,2);not null" json:"rate"`
	Amount     float64   `gorm:"type:decimal(10,2);not null" json:"amount"`
	Status     string    `gorm:"size:20;default:'pending'" json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`

	// Relations
	Agent Agent `gorm:"foreignKey:AgentID" json:"agent,omitempty"`
}

func (Commission) TableName() string {
	return "commissions"
}

// CalculateCommission calculates commission amount from order total and rate
func CalculateCommission(orderTotal, rate float64) float64 {
	return (orderTotal * rate) / 100.0
}
