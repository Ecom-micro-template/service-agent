package persistence

import (
	"context"

	"github.com/Ecom-micro-template/service-agent/internal/domain"
	"gorm.io/gorm"
)

// CategoryCommissionRepository defines the interface for agent category commission data operations
type CategoryCommissionRepository interface {
	GetByAgentID(ctx context.Context, agentID uint) ([]domain.AgentCategoryCommission, error)
	DeleteByAgentID(ctx context.Context, agentID uint) error
	Create(ctx context.Context, commission *domain.AgentCategoryCommission) error
	BulkReplace(ctx context.Context, agentID uint, commissions []domain.AgentCategoryCommission) error
}

// categoryCommissionRepository implements CategoryCommissionRepository
type categoryCommissionRepository struct {
	db *gorm.DB
}

// NewCategoryCommissionRepository creates a new category commission repository
func NewCategoryCommissionRepository(db *gorm.DB) CategoryCommissionRepository {
	return &categoryCommissionRepository{db: db}
}

// GetByAgentID retrieves all category commissions for an agent
func (r *categoryCommissionRepository) GetByAgentID(ctx context.Context, agentID uint) ([]domain.AgentCategoryCommission, error) {
	var commissions []domain.AgentCategoryCommission
	err := r.db.WithContext(ctx).Where("agent_id = ?", agentID).Find(&commissions).Error
	return commissions, err
}

// DeleteByAgentID deletes all category commissions for an agent
func (r *categoryCommissionRepository) DeleteByAgentID(ctx context.Context, agentID uint) error {
	return r.db.WithContext(ctx).Where("agent_id = ?", agentID).Delete(&domain.AgentCategoryCommission{}).Error
}

// Create creates a new category commission
func (r *categoryCommissionRepository) Create(ctx context.Context, commission *domain.AgentCategoryCommission) error {
	return r.db.WithContext(ctx).Create(commission).Error
}

// BulkReplace replaces all category commissions for an agent
func (r *categoryCommissionRepository) BulkReplace(ctx context.Context, agentID uint, commissions []domain.AgentCategoryCommission) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Delete existing
		if err := tx.Where("agent_id = ?", agentID).Delete(&domain.AgentCategoryCommission{}).Error; err != nil {
			return err
		}

		// Create new ones
		for i := range commissions {
			commissions[i].AgentID = agentID
			if err := tx.Create(&commissions[i]).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
