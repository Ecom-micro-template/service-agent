package commission

import (
	"github.com/niaga-platform/service-agent/internal/domain/shared"
)

// Calculator is a domain service for calculating commissions.
type Calculator struct{}

// NewCalculator creates a new commission calculator.
func NewCalculator() *Calculator {
	return &Calculator{}
}

// CalculationParams contains parameters for commission calculation.
type CalculationParams struct {
	OrderTotal     float64
	BaseRate       shared.CommissionRate
	AgentTier      shared.AgentTier
	CategoryRates  map[string]shared.CommissionRate // Optional category-specific rates
	ProductAmounts map[string]float64               // productID -> amount
}

// CalculationResult contains the result of commission calculation.
type CalculationResult struct {
	BaseAmount    float64
	TierBonus     float64
	TotalAmount   float64
	EffectiveRate float64
}

// Calculate calculates commission with tier bonuses.
func (c *Calculator) Calculate(params CalculationParams) CalculationResult {
	// Base commission
	baseAmount := params.BaseRate.CalculateCommission(params.OrderTotal)

	// Tier bonus
	tierBonusRate := params.AgentTier.BonusPercentage()
	tierBonus := params.OrderTotal * tierBonusRate

	totalAmount := baseAmount + tierBonus
	effectiveRate := 0.0
	if params.OrderTotal > 0 {
		effectiveRate = (totalAmount / params.OrderTotal) * 100
	}

	return CalculationResult{
		BaseAmount:    baseAmount,
		TierBonus:     tierBonus,
		TotalAmount:   totalAmount,
		EffectiveRate: effectiveRate,
	}
}

// CalculateSimple calculates commission without tier bonus.
func (c *Calculator) CalculateSimple(orderTotal float64, rate shared.CommissionRate) float64 {
	return rate.CalculateCommission(orderTotal)
}

// CalculateWithTier calculates commission with tier bonus applied.
func (c *Calculator) CalculateWithTier(orderTotal float64, rate shared.CommissionRate, tier shared.AgentTier) float64 {
	baseAmount := rate.CalculateCommission(orderTotal)
	tierBonus := orderTotal * tier.BonusPercentage()
	return baseAmount + tierBonus
}

// CalculateCategoryBased calculates commission with category-specific rates.
func (c *Calculator) CalculateCategoryBased(
	productAmounts map[string]float64, // categoryID -> amount
	categoryRates map[string]shared.CommissionRate,
	defaultRate shared.CommissionRate,
	tier shared.AgentTier,
) float64 {
	var totalCommission float64
	var totalOrder float64

	for categoryID, amount := range productAmounts {
		rate := defaultRate
		if r, exists := categoryRates[categoryID]; exists {
			rate = r
		}
		totalCommission += rate.CalculateCommission(amount)
		totalOrder += amount
	}

	// Add tier bonus on total
	tierBonus := totalOrder * tier.BonusPercentage()

	return totalCommission + tierBonus
}

// EffectiveRate calculates the effective rate including tier bonus.
func (c *Calculator) EffectiveRate(baseRate shared.CommissionRate, tier shared.AgentTier) shared.CommissionRate {
	return baseRate.AddPercentage(tier.BonusPercentage())
}
