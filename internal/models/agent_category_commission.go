// Package models contains GORM persistence models for the agent service.
//
// Deprecated: This package is being migrated to DDD architecture.
// For new development, use:
//   - Domain models: github.com/Ecom-micro-template/service-agent/internal/domain/commission
//   - Persistence: github.com/Ecom-micro-template/service-agent/internal/infrastructure/persistence
//
// Existing code can continue using this package during the transition period.
package models

import (
	"time"

	"gorm.io/gorm"
)

// AgentCategoryCommission stores category-specific commission rates per agent
type AgentCategoryCommission struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	AgentID        uint           `gorm:"index;not null" json:"agent_id"`
	CategoryID     string         `gorm:"size:50;not null" json:"category_id"` // UUID from catalog service
	CategoryName   string         `gorm:"size:255" json:"category_name"`       // Cached for display
	CommissionRate float64        `gorm:"type:decimal(5,2);not null" json:"commission_rate"`
	IsActive       bool           `gorm:"default:true" json:"is_active"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Agent Agent `gorm:"foreignKey:AgentID" json:"agent,omitempty"`
}

func (AgentCategoryCommission) TableName() string {
	return "agent_category_commissions"
}

// UpdateCategoryCommissionsRequest for bulk updating category commissions
type UpdateCategoryCommissionsRequest struct {
	Commissions []CategoryCommissionInput `json:"commissions" binding:"required"`
}

// CategoryCommissionInput represents a single category commission setting
type CategoryCommissionInput struct {
	CategoryID     string  `json:"category_id" binding:"required"`
	CategoryName   string  `json:"category_name"`
	CommissionRate float64 `json:"commission_rate" binding:"required,min=0,max=100"`
	IsActive       bool    `json:"is_active"`
}
