package persistence

import (
	"time"
)

// TeamModel is the GORM persistence model for Team.
type TeamModel struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	Code           string    `gorm:"uniqueIndex;size:50;not null" json:"code"`
	Name           string    `gorm:"size:255;not null" json:"name"`
	Description    string    `gorm:"type:text" json:"description"`
	LeaderID       *uint     `gorm:"index" json:"leader_id,omitempty"`
	TargetMonthly  float64   `gorm:"type:decimal(12,2);default:0" json:"target_monthly"`
	CommissionRate float64   `gorm:"type:decimal(5,2);default:10.0" json:"commission_rate"`
	IsActive       bool      `gorm:"default:true" json:"is_active"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`

	// Relations
	Leader  *AgentModel  `gorm:"foreignKey:LeaderID" json:"leader,omitempty"`
	Members []AgentModel `gorm:"foreignKey:TeamID" json:"members,omitempty"`
}

// TableName specifies the table name.
func (TeamModel) TableName() string {
	return "teams"
}
