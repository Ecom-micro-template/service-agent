package handlers

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/Ecom-micro-template/lib-common-go/health"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	checker *health.Checker
	handler *health.GinHandler
	logger  *zap.Logger
}

// HealthDependencies contains all dependencies to be health-checked
type HealthDependencies struct {
	DB *gorm.DB
}

// NewHealthHandler creates a new health handler with all dependency checks
func NewHealthHandler(
	serviceName string,
	version string,
	deps HealthDependencies,
	logger *zap.Logger,
) *HealthHandler {
	checker := health.NewChecker(serviceName, version, logger)
	checker.SetTimeout(3 * time.Second)

	// Register PostgreSQL check
	if deps.DB != nil {
		checker.RegisterCheck("postgres", health.PostgresCheck(deps.DB))
	}

	return &HealthHandler{
		checker: checker,
		handler: health.NewGinHandler(checker),
		logger:  logger,
	}
}

// RegisterRoutes registers health check routes
func (h *HealthHandler) RegisterRoutes(router gin.IRouter) {
	h.handler.RegisterRoutes(router)
}

// Health returns comprehensive health status
func (h *HealthHandler) Health(c *gin.Context) {
	h.handler.Health(c)
}

// Liveness returns simple liveness status
func (h *HealthHandler) Liveness(c *gin.Context) {
	h.handler.Liveness(c)
}

// Readiness returns readiness status
func (h *HealthHandler) Readiness(c *gin.Context) {
	h.handler.Readiness(c)
}

// RegisterCustomCheck allows adding custom health checks
func (h *HealthHandler) RegisterCustomCheck(name string, check health.CheckFunc) {
	h.checker.RegisterCheck(name, check)
}
