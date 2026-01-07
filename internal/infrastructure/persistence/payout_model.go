package persistence

import (
	"time"
)

// PayoutModel is the GORM persistence model for Payout.
type PayoutModel struct {
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
	Agent AgentModel `gorm:"foreignKey:AgentID" json:"agent,omitempty"`
}

// TableName specifies the table name.
func (PayoutModel) TableName() string {
	return "payouts"
}
