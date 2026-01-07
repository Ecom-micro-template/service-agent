// Package persistence contains GORM models and repository implementations.
package persistence

import (
	"time"

	"gorm.io/gorm"
)

// AgentModel is the GORM persistence model for Agent.
type AgentModel struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	Code           string    `gorm:"uniqueIndex;size:50;not null" json:"code"`
	Name           string    `gorm:"size:255;not null" json:"name"`
	Email          string    `gorm:"uniqueIndex;size:255;not null" json:"email"`
	Phone          string    `gorm:"size:50" json:"phone"`
	CommissionRate float64   `gorm:"type:decimal(5,2);default:10.0" json:"commission_rate"`
	Tier           string    `gorm:"size:20;default:'bronze'" json:"tier"`
	Status         string    `gorm:"size:20;default:'active'" json:"status"`
	TotalEarned    float64   `gorm:"type:decimal(10,2);default:0" json:"total_earned"`
	TeamID         *uint     `gorm:"index" json:"team_id,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`

	// Relations
	Commissions []CommissionModel `gorm:"foreignKey:AgentID" json:"commissions,omitempty"`
	Payouts     []PayoutModel     `gorm:"foreignKey:AgentID" json:"payouts,omitempty"`
	Team        *TeamModel        `gorm:"foreignKey:TeamID" json:"team,omitempty"`
}

// TableName specifies the table name.
func (AgentModel) TableName() string {
	return "agents"
}

// BeforeCreate hook to set defaults.
func (m *AgentModel) BeforeCreate(tx *gorm.DB) error {
	if m.Status == "" {
		m.Status = "active"
	}
	if m.Tier == "" {
		m.Tier = "bronze"
	}
	if m.CommissionRate == 0 {
		m.CommissionRate = 10.0
	}
	return nil
}
