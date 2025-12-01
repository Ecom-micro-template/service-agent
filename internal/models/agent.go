package models

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

type Agent struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	Code           string    `gorm:"uniqueIndex;size:50;not null" json:"code"`
	Name           string    `gorm:"size:255;not null" json:"name"`
	Email          string    `gorm:"uniqueIndex;size:255;not null" json:"email"`
	Phone          string    `gorm:"size:50" json:"phone"`
	CommissionRate float64   `gorm:"type:decimal(5,2);default:10.0" json:"commission_rate"`
	Status         string    `gorm:"size:20;default:'active'" json:"status"`
	TotalEarned    float64   `gorm:"type:decimal(10,2);default:0" json:"total_earned"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`

	// Relations
	Commissions []Commission `gorm:"foreignKey:AgentID" json:"commissions,omitempty"`
	Payouts     []Payout     `gorm:"foreignKey:AgentID" json:"payouts,omitempty"`
}

func (a *Agent) BeforeCreate(tx *gorm.DB) error {
	if a.Code == "" {
		// Generate agent code: AGT + ID padded to 4 digits
		var count int64
		tx.Model(&Agent{}).Count(&count)
		a.Code = fmt.Sprintf("AGT%04d", count+1)
	}
	if a.Status == "" {
		a.Status = "active"
	}
	if a.CommissionRate == 0 {
		a.CommissionRate = 10.0
	}
	return nil
}

func (Agent) TableName() string {
	return "agents"
}
