package persistence

import (
	"time"
)

// CommissionModel is the GORM persistence model for Commission.
type CommissionModel struct {
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
	Agent AgentModel `gorm:"foreignKey:AgentID" json:"agent,omitempty"`
}

// TableName specifies the table name.
func (CommissionModel) TableName() string {
	return "commissions"
}
