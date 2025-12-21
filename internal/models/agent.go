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
	TeamID         *uint     `gorm:"index" json:"team_id,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`

	// Relations
	Commissions []Commission `gorm:"foreignKey:AgentID" json:"commissions,omitempty"`
	Payouts     []Payout     `gorm:"foreignKey:AgentID" json:"payouts,omitempty"`
	Team        *Team        `gorm:"foreignKey:TeamID" json:"team,omitempty"`
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

// Dashboard represents agent dashboard statistics
type Dashboard struct {
	TotalOrders         int64               `json:"total_orders"`
	TotalSales          float64             `json:"total_sales"`
	TotalCommission     float64             `json:"total_commission"`
	PendingCommission   float64             `json:"pending_commission"`
	ApprovedCommission  float64             `json:"approved_commission"`
	PaidCommission      float64             `json:"paid_commission"`
	TotalCustomers      int64               `json:"total_customers"`
	MonthlyOrders       int64               `json:"monthly_orders"`
	MonthlySales        float64             `json:"monthly_sales"`
	MonthlyCommission   float64             `json:"monthly_commission"`
	AverageOrderValue   float64             `json:"average_order_value"`
	CommissionBreakdown CommissionBreakdown `json:"commission_breakdown"`
}

// CommissionBreakdown shows commission by status
type CommissionBreakdown struct {
	Pending  float64 `json:"pending"`
	Approved float64 `json:"approved"`
	Paid     float64 `json:"paid"`
}

// Customer represents an agent's customer
type Customer struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	AgentID     *uint      `gorm:"index" json:"agent_id,omitempty"`
	Name        string     `gorm:"size:255;not null" json:"name"`
	Email       string     `gorm:"uniqueIndex;size:255;not null" json:"email"`
	Phone       string     `gorm:"size:50" json:"phone"`
	Address     string     `gorm:"type:text" json:"address"`
	City        string     `gorm:"size:100" json:"city"`
	State       string     `gorm:"size:100" json:"state"`
	Postcode    string     `gorm:"size:20" json:"postcode"`
	TotalOrders int        `gorm:"default:0" json:"total_orders"`
	TotalSpent  float64    `gorm:"type:decimal(12,2);default:0" json:"total_spent"`
	LastOrderAt *time.Time `json:"last_order_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func (Customer) TableName() string {
	return "customers"
}

// Order represents an agent's order (simplified view)
type Order struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	OrderNumber    string    `gorm:"uniqueIndex;size:50;not null" json:"order_number"`
	AgentID        uint      `gorm:"index;not null" json:"agent_id"`
	CustomerID     uint      `gorm:"index;not null" json:"customer_id"`
	CustomerName   string    `gorm:"size:255" json:"customer_name"`
	CustomerEmail  string    `gorm:"size:255" json:"customer_email"`
	Total          float64   `gorm:"type:decimal(12,2);not null" json:"total"`
	Status         string    `gorm:"size:20;default:'pending'" json:"status"`
	PaymentStatus  string    `gorm:"size:20;default:'unpaid'" json:"payment_status"`
	CommissionRate float64   `gorm:"type:decimal(5,2)" json:"commission_rate"`
	Commission     float64   `gorm:"type:decimal(10,2)" json:"commission"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (Order) TableName() string {
	return "orders"
}

// Team represents a sales team
type Team struct {
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
	Leader  *Agent  `gorm:"foreignKey:LeaderID" json:"leader,omitempty"`
	Members []Agent `gorm:"foreignKey:TeamID" json:"members,omitempty"`
}

func (Team) TableName() string {
	return "teams"
}

// Performance represents monthly performance metrics
type Performance struct {
	Month              time.Time `json:"month"`
	TotalSales         float64   `json:"total_sales"`
	TotalOrders        int64     `json:"total_orders"`
	TotalCommission    float64   `json:"total_commission"`
	CommissionPending  float64   `json:"commission_pending"`
	CommissionApproved float64   `json:"commission_approved"`
	CommissionPaid     float64   `json:"commission_paid"`
}

// CreateCustomerRequest represents customer creation payload
type CreateCustomerRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Phone    string `json:"phone"`
	Address  string `json:"address"`
	City     string `json:"city"`
	State    string `json:"state"`
	Postcode string `json:"postcode"`
}

// UpdateCustomerRequest represents customer update payload
type UpdateCustomerRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email" binding:"omitempty,email"`
	Phone    string `json:"phone"`
	Address  string `json:"address"`
	City     string `json:"city"`
	State    string `json:"state"`
	Postcode string `json:"postcode"`
}

// CreateOrderRequest represents order creation payload
type CreateOrderRequest struct {
	CustomerID uint        `json:"customer_id" binding:"required"`
	Items      []OrderItem `json:"items" binding:"required,min=1"`
	Notes      string      `json:"notes"`
}

// OrderItem represents an item in the order
type OrderItem struct {
	ProductID uint    `json:"product_id" binding:"required"`
	Quantity  int     `json:"quantity" binding:"required,min=1"`
	Price     float64 `json:"price" binding:"required,min=0"`
}
