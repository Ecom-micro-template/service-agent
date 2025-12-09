package services

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// CommissionStatus represents commission status
type CommissionStatus string

const (
	CommissionStatusPending  CommissionStatus = "pending"
	CommissionStatusApproved CommissionStatus = "approved"
	CommissionStatusPaid     CommissionStatus = "paid"
	CommissionStatusRejected CommissionStatus = "rejected"
)

// CommissionTier represents tiered commission structure
type CommissionTier struct {
	MinAmount float64
	MaxAmount float64
	Rate      float64
}

// CommissionCalculationRequest represents a request to calculate commission
type CommissionCalculationRequest struct {
	OrderID        uuid.UUID
	AgentID        uuid.UUID
	OrderTotal     float64
	OrderSubtotal  float64
	ShippingCost   float64
	DiscountAmount float64
	ProductIDs     []uuid.UUID
	CategoryIDs    []uuid.UUID
}

// CommissionCalculationResult represents calculated commission
type CommissionCalculationResult struct {
	OrderID           uuid.UUID
	AgentID           uuid.UUID
	CommissionRate    float64
	CommissionAmount  float64
	BasedOnAmount     float64
	TierApplied       *CommissionTier
	ProductCommission map[uuid.UUID]float64
	Breakdown         []CommissionBreakdownItem
}

// CommissionBreakdownItem shows commission per item/category
type CommissionBreakdownItem struct {
	ItemType string  // "product", "category", "base"
	ItemID   string
	ItemName string
	Amount   float64
	Rate     float64
}

// AgentCommission represents a commission record
type AgentCommission struct {
	ID               uuid.UUID        `gorm:"type:uuid;primary_key"`
	AgentID          uuid.UUID        `gorm:"type:uuid;not null;index"`
	OrderID          uuid.UUID        `gorm:"type:uuid;not null;index"`
	OrderTotal       float64          `gorm:"type:decimal(12,2)"`
	CommissionRate   float64          `gorm:"type:decimal(5,2)"`
	CommissionAmount float64          `gorm:"type:decimal(10,2)"`
	Status           CommissionStatus `gorm:"type:varchar(20);default:'pending'"`
	BasedOnAmount    float64          `gorm:"type:decimal(12,2)"` // Amount used for calculation
	Notes            string           `gorm:"type:text"`
	ApprovedBy       *uuid.UUID       `gorm:"type:uuid"`
	ApprovedAt       *time.Time
	PaidAt           *time.Time
	RejectedAt       *time.Time
	RejectionReason  string `gorm:"type:text"`
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func (AgentCommission) TableName() string {
	return "sales.agent_commissions"
}

// CommissionCalculatorService handles commission calculations
type CommissionCalculatorService struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewCommissionCalculatorService creates a new commission calculator
func NewCommissionCalculatorService(db *gorm.DB, logger *zap.Logger) *CommissionCalculatorService {
	return &CommissionCalculatorService{
		db:     db,
		logger: logger,
	}
}

// CalculateCommission calculates commission for an order
func (s *CommissionCalculatorService) CalculateCommission(req *CommissionCalculationRequest) (*CommissionCalculationResult, error) {
	result := &CommissionCalculationResult{
		OrderID:           req.OrderID,
		AgentID:           req.AgentID,
		ProductCommission: make(map[uuid.UUID]float64),
		Breakdown:         []CommissionBreakdownItem{},
	}

	// Get agent's commission rate and tier
	agent, err := s.getAgentCommissionRate(req.AgentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent commission rate: %w", err)
	}

	// Determine base amount for commission (typically subtotal, excluding shipping/discounts)
	baseAmount := req.OrderSubtotal

	// Check if agent has tiered commission
	tier := s.getTierForAmount(agent.TierRates, baseAmount)
	if tier != nil {
		result.CommissionRate = tier.Rate
		result.TierApplied = tier
	} else {
		result.CommissionRate = agent.BaseRate
	}

	// Calculate base commission
	commissionAmount := baseAmount * (result.CommissionRate / 100)

	// Add product-specific commissions
	if len(req.ProductIDs) > 0 {
		productBonus, breakdown := s.calculateProductCommissions(req.ProductIDs, req.OrderSubtotal, agent.BaseRate)
		commissionAmount += productBonus
		result.Breakdown = append(result.Breakdown, breakdown...)
	}

	// Add category-specific commissions
	if len(req.CategoryIDs) > 0 {
		categoryBonus, breakdown := s.calculateCategoryCommissions(req.CategoryIDs, req.OrderSubtotal, agent.BaseRate)
		commissionAmount += categoryBonus
		result.Breakdown = append(result.Breakdown, breakdown...)
	}

	result.CommissionAmount = commissionAmount
	result.BasedOnAmount = baseAmount

	s.logger.Info("Commission calculated",
		zap.String("agent_id", req.AgentID.String()),
		zap.String("order_id", req.OrderID.String()),
		zap.Float64("commission", commissionAmount),
		zap.Float64("rate", result.CommissionRate),
	)

	return result, nil
}

// getAgentCommissionRate gets agent's commission configuration
func (s *CommissionCalculatorService) getAgentCommissionRate(agentID uuid.UUID) (*AgentCommissionConfig, error) {
	var agent struct {
		ID           uuid.UUID `gorm:"column:id"`
		BaseRate     float64   `gorm:"column:commission_rate"`
		TierEnabled  bool      `gorm:"column:tier_enabled"`
		TeamID       *uuid.UUID `gorm:"column:team_id"`
		IsActive     bool      `gorm:"column:is_active"`
	}

	if err := s.db.Table("sales.agents").
		Where("id = ?", agentID).
		First(&agent).Error; err != nil {
		return nil, fmt.Errorf("agent not found: %w", err)
	}

	if !agent.IsActive {
		return nil, fmt.Errorf("agent is not active")
	}

	config := &AgentCommissionConfig{
		AgentID:     agent.ID,
		BaseRate:    agent.BaseRate,
		TierEnabled: agent.TierEnabled,
		TierRates:   []CommissionTier{},
	}

	// Load tiered rates if enabled
	if agent.TierEnabled {
		var tiers []struct {
			MinAmount float64 `gorm:"column:min_amount"`
			MaxAmount float64 `gorm:"column:max_amount"`
			Rate      float64 `gorm:"column:rate"`
		}

		s.db.Table("sales.commission_tiers").
			Where("agent_id = ?", agentID).
			Order("min_amount ASC").
			Find(&tiers)

		for _, tier := range tiers {
			config.TierRates = append(config.TierRates, CommissionTier{
				MinAmount: tier.MinAmount,
				MaxAmount: tier.MaxAmount,
				Rate:      tier.Rate,
			})
		}
	}

	// Check team commission boost
	if agent.TeamID != nil {
		var teamBoost float64
		s.db.Table("sales.teams").
			Select("commission_boost").
			Where("id = ?", agent.TeamID).
			Scan(&teamBoost)

		if teamBoost > 0 {
			config.BaseRate += teamBoost
		}
	}

	return config, nil
}

// getTierForAmount returns the commission tier for given amount
func (s *CommissionCalculatorService) getTierForAmount(tiers []CommissionTier, amount float64) *CommissionTier {
	for _, tier := range tiers {
		if amount >= tier.MinAmount && (tier.MaxAmount == 0 || amount <= tier.MaxAmount) {
			return &tier
		}
	}
	return nil
}

// calculateProductCommissions calculates bonus commissions for specific products
func (s *CommissionCalculatorService) calculateProductCommissions(productIDs []uuid.UUID, orderSubtotal, baseRate float64) (float64, []CommissionBreakdownItem) {
	var bonusCommission float64
	breakdown := []CommissionBreakdownItem{}

	// Check if products have special commission rates
	var productRates []struct {
		ProductID      uuid.UUID `gorm:"column:product_id"`
		ProductName    string    `gorm:"column:product_name"`
		BonusRate      float64   `gorm:"column:commission_bonus_rate"`
	}

	s.db.Table("sales.product_commission_rates").
		Select("product_id, product_name, commission_bonus_rate").
		Where("product_id IN ? AND is_active = ?", productIDs, true).
		Find(&productRates)

	for _, pr := range productRates {
		bonus := orderSubtotal * (pr.BonusRate / 100)
		bonusCommission += bonus

		breakdown = append(breakdown, CommissionBreakdownItem{
			ItemType: "product",
			ItemID:   pr.ProductID.String(),
			ItemName: pr.ProductName,
			Amount:   bonus,
			Rate:     pr.BonusRate,
		})
	}

	return bonusCommission, breakdown
}

// calculateCategoryCommissions calculates bonus commissions for categories
func (s *CommissionCalculatorService) calculateCategoryCommissions(categoryIDs []uuid.UUID, orderSubtotal, baseRate float64) (float64, []CommissionBreakdownItem) {
	var bonusCommission float64
	breakdown := []CommissionBreakdownItem{}

	// Check if categories have special commission rates
	var categoryRates []struct {
		CategoryID  uuid.UUID `gorm:"column:category_id"`
		CategoryName string   `gorm:"column:category_name"`
		BonusRate   float64   `gorm:"column:commission_bonus_rate"`
	}

	s.db.Table("sales.category_commission_rates").
		Select("category_id, category_name, commission_bonus_rate").
		Where("category_id IN ? AND is_active = ?", categoryIDs, true).
		Find(&categoryRates)

	for _, cr := range categoryRates {
		bonus := orderSubtotal * (cr.BonusRate / 100)
		bonusCommission += bonus

		breakdown = append(breakdown, CommissionBreakdownItem{
			ItemType: "category",
			ItemID:   cr.CategoryID.String(),
			ItemName: cr.CategoryName,
			Amount:   bonus,
			Rate:     cr.BonusRate,
		})
	}

	return bonusCommission, breakdown
}

// CreateCommission creates a commission record
func (s *CommissionCalculatorService) CreateCommission(result *CommissionCalculationResult) error {
	commission := &AgentCommission{
		ID:               uuid.New(),
		AgentID:          result.AgentID,
		OrderID:          result.OrderID,
		OrderTotal:       0, // Will be set by caller
		CommissionRate:   result.CommissionRate,
		CommissionAmount: result.CommissionAmount,
		BasedOnAmount:    result.BasedOnAmount,
		Status:           CommissionStatusPending,
	}

	if err := s.db.Create(commission).Error; err != nil {
		return fmt.Errorf("failed to create commission: %w", err)
	}

	s.logger.Info("Commission created",
		zap.String("commission_id", commission.ID.String()),
		zap.String("agent_id", result.AgentID.String()),
		zap.Float64("amount", result.CommissionAmount),
	)

	return nil
}

// ApproveCommission approves a commission
func (s *CommissionCalculatorService) ApproveCommission(commissionID, approverID uuid.UUID, notes string) error {
	now := time.Now()
	updates := map[string]interface{}{
		"status":      CommissionStatusApproved,
		"approved_by": approverID,
		"approved_at": now,
		"notes":       notes,
	}

	if err := s.db.Table("sales.agent_commissions").
		Where("id = ? AND status = ?", commissionID, CommissionStatusPending).
		Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to approve commission: %w", err)
	}

	// Update agent's total earned
	var commission AgentCommission
	if err := s.db.First(&commission, commissionID).Error; err == nil {
		s.db.Table("sales.agents").
			Where("id = ?", commission.AgentID).
			Update("total_earned", gorm.Expr("total_earned + ?", commission.CommissionAmount))
	}

	s.logger.Info("Commission approved",
		zap.String("commission_id", commissionID.String()),
		zap.String("approver_id", approverID.String()),
	)

	return nil
}

// RejectCommission rejects a commission
func (s *CommissionCalculatorService) RejectCommission(commissionID, approverID uuid.UUID, reason string) error {
	now := time.Now()
	updates := map[string]interface{}{
		"status":           CommissionStatusRejected,
		"rejected_at":      now,
		"rejection_reason": reason,
		"approved_by":      approverID, // Track who rejected
	}

	if err := s.db.Table("sales.agent_commissions").
		Where("id = ? AND status = ?", commissionID, CommissionStatusPending).
		Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to reject commission: %w", err)
	}

	s.logger.Info("Commission rejected",
		zap.String("commission_id", commissionID.String()),
		zap.String("reason", reason),
	)

	return nil
}

// MarkCommissionPaid marks a commission as paid
func (s *CommissionCalculatorService) MarkCommissionPaid(commissionID uuid.UUID, paidAt time.Time) error {
	updates := map[string]interface{}{
		"status":  CommissionStatusPaid,
		"paid_at": paidAt,
	}

	if err := s.db.Table("sales.agent_commissions").
		Where("id = ? AND status = ?", commissionID, CommissionStatusApproved).
		Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to mark commission as paid: %w", err)
	}

	s.logger.Info("Commission marked as paid",
		zap.String("commission_id", commissionID.String()),
	)

	return nil
}

// GetCommissionsByAgent gets all commissions for an agent
func (s *CommissionCalculatorService) GetCommissionsByAgent(agentID uuid.UUID, status *CommissionStatus, limit, offset int) ([]AgentCommission, int64, error) {
	var commissions []AgentCommission
	var total int64

	query := s.db.Table("sales.agent_commissions").Where("agent_id = ?", agentID)

	if status != nil {
		query = query.Where("status = ?", *status)
	}

	// Get total count
	query.Count(&total)

	// Get paginated results
	if err := query.
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&commissions).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get commissions: %w", err)
	}

	return commissions, total, nil
}

// GetCommissionStats gets commission statistics for an agent
func (s *CommissionCalculatorService) GetCommissionStats(agentID uuid.UUID) (*CommissionStats, error) {
	stats := &CommissionStats{
		AgentID: agentID,
	}

	// Get total pending
	s.db.Table("sales.agent_commissions").
		Select("COALESCE(SUM(commission_amount), 0)").
		Where("agent_id = ? AND status = ?", agentID, CommissionStatusPending).
		Scan(&stats.TotalPending)

	// Get total approved
	s.db.Table("sales.agent_commissions").
		Select("COALESCE(SUM(commission_amount), 0)").
		Where("agent_id = ? AND status = ?", agentID, CommissionStatusApproved).
		Scan(&stats.TotalApproved)

	// Get total paid
	s.db.Table("sales.agent_commissions").
		Select("COALESCE(SUM(commission_amount), 0)").
		Where("agent_id = ? AND status = ?", agentID, CommissionStatusPaid).
		Scan(&stats.TotalPaid)

	// Get count
	s.db.Table("sales.agent_commissions").
		Where("agent_id = ?", agentID).
		Count(&stats.TotalCount)

	stats.TotalEarned = stats.TotalPaid
	stats.TotalUnpaid = stats.TotalPending + stats.TotalApproved

	return stats, nil
}

// AgentCommissionConfig holds agent commission configuration
type AgentCommissionConfig struct {
	AgentID     uuid.UUID
	BaseRate    float64
	TierEnabled bool
	TierRates   []CommissionTier
}

// CommissionStats holds commission statistics
type CommissionStats struct {
	AgentID       uuid.UUID
	TotalPending  float64
	TotalApproved float64
	TotalPaid     float64
	TotalEarned   float64
	TotalUnpaid   float64
	TotalCount    int64
}

// CalculateMonthlyPerformance calculates monthly commission performance
func (s *CommissionCalculatorService) CalculateMonthlyPerformance(agentID uuid.UUID, year int, month int) (*MonthlyPerformance, error) {
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, 0)

	perf := &MonthlyPerformance{
		AgentID: agentID,
		Year:    year,
		Month:   month,
	}

	// Get commission summary
	var summary struct {
		TotalCommission float64
		TotalOrders     int64
		TotalSales      float64
	}

	s.db.Table("sales.agent_commissions").
		Select(`
			COALESCE(SUM(commission_amount), 0) as total_commission,
			COUNT(*) as total_orders,
			COALESCE(SUM(order_total), 0) as total_sales
		`).
		Where("agent_id = ? AND created_at >= ? AND created_at < ?", agentID, startDate, endDate).
		Scan(&summary)

	perf.TotalCommission = summary.TotalCommission
	perf.TotalOrders = summary.TotalOrders
	perf.TotalSales = summary.TotalSales

	if perf.TotalOrders > 0 {
		perf.AverageOrderValue = perf.TotalSales / float64(perf.TotalOrders)
		perf.AverageCommission = perf.TotalCommission / float64(perf.TotalOrders)
	}

	return perf, nil
}

// MonthlyPerformance represents monthly commission performance
type MonthlyPerformance struct {
	AgentID           uuid.UUID
	Year              int
	Month             int
	TotalSales        float64
	TotalOrders       int64
	TotalCommission   float64
	AverageOrderValue float64
	AverageCommission float64
}
