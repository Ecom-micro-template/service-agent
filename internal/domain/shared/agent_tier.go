package shared

import (
	"errors"
	"fmt"
)

// AgentTier represents the tier/level of an agent.
type AgentTier string

// Agent tier constants
const (
	TierBronze   AgentTier = "bronze"
	TierSilver   AgentTier = "silver"
	TierGold     AgentTier = "gold"
	TierPlatinum AgentTier = "platinum"
)

// Tier bonus percentages
const (
	BonusBronze   = 0.00
	BonusSilver   = 0.01 // 1%
	BonusGold     = 0.02 // 2%
	BonusPlatinum = 0.03 // 3%
)

// ErrInvalidAgentTier is returned for invalid tier values.
var ErrInvalidAgentTier = errors.New("invalid agent tier")

// AllAgentTiers returns all valid tiers in order.
func AllAgentTiers() []AgentTier {
	return []AgentTier{TierBronze, TierSilver, TierGold, TierPlatinum}
}

// IsValid returns true if the tier is valid.
func (t AgentTier) IsValid() bool {
	switch t {
	case TierBronze, TierSilver, TierGold, TierPlatinum:
		return true
	default:
		return false
	}
}

// String returns the string representation.
func (t AgentTier) String() string {
	return string(t)
}

// Label returns a human-readable label.
func (t AgentTier) Label() string {
	switch t {
	case TierBronze:
		return "Bronze"
	case TierSilver:
		return "Silver"
	case TierGold:
		return "Gold"
	case TierPlatinum:
		return "Platinum"
	default:
		return "Unknown"
	}
}

// BonusPercentage returns the bonus commission percentage for this tier.
func (t AgentTier) BonusPercentage() float64 {
	switch t {
	case TierPlatinum:
		return BonusPlatinum
	case TierGold:
		return BonusGold
	case TierSilver:
		return BonusSilver
	case TierBronze:
		return BonusBronze
	default:
		return 0.0
	}
}

// Level returns the numeric level (1-4).
func (t AgentTier) Level() int {
	switch t {
	case TierPlatinum:
		return 4
	case TierGold:
		return 3
	case TierSilver:
		return 2
	case TierBronze:
		return 1
	default:
		return 0
	}
}

// IsHigherThan returns true if this tier is higher than other.
func (t AgentTier) IsHigherThan(other AgentTier) bool {
	return t.Level() > other.Level()
}

// NextTier returns the next tier (or same if already max).
func (t AgentTier) NextTier() AgentTier {
	switch t {
	case TierBronze:
		return TierSilver
	case TierSilver:
		return TierGold
	case TierGold:
		return TierPlatinum
	default:
		return t
	}
}

// IsPremium returns true if tier is gold or higher.
func (t AgentTier) IsPremium() bool {
	return t == TierGold || t == TierPlatinum
}

// ParseAgentTier parses a string into an AgentTier.
func ParseAgentTier(s string) (AgentTier, error) {
	tier := AgentTier(s)
	if !tier.IsValid() {
		return "", fmt.Errorf("%w: %s", ErrInvalidAgentTier, s)
	}
	return tier, nil
}
