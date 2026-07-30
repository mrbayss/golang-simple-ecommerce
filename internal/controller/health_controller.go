package controller

import (
	"github.com/gofiber/fiber/v3"
	"github.com/mrbayss/golang-simple-ecommerce/internal/model"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type HealthCheckResult struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

type HealthResponse struct {
	Status string                `json:"status"`
	Checks []HealthCheckResult   `json:"checks"`
}

type HealthController struct {
	Log    *logrus.Logger
	DB     *gorm.DB
	Redis  *redis.Client
	Config interface{}
}

func NewHealthController(log *logrus.Logger, db *gorm.DB, redis *redis.Client) *HealthController {
	return &HealthController{
		Log:   log,
		DB:    db,
		Redis: redis,
	}
}

func (hc *HealthController) Check(ctx fiber.Ctx) error {
	var checks []HealthCheckResult

	// DB check
	dbStatus := "ok"
	var dbErr error
	if err := hc.DB.Raw("SELECT 1").Error; err != nil {
		dbStatus = "error"
		dbErr = err
		hc.Log.Warnf("health check db failed: %v", err)
	}
	checks = append(checks, HealthCheckResult{
		Name:   "database",
		Status: dbStatus,
		Error:  errorToString(dbErr),
	})

	// Redis check
	redisStatus := "ok"
	var redisErr error
	if err := hc.Redis.Ping(ctx.Context()).Err(); err != nil {
		redisStatus = "error"
		redisErr = err
		hc.Log.Warnf("health check redis failed: %v", err)
	}
	checks = append(checks, HealthCheckResult{
		Name:   "redis",
		Status: redisStatus,
		Error:  errorToString(redisErr),
	})

	// overall status
	overall := "healthy"
	for _, c := range checks {
		if c.Status != "ok" {
			overall = "unhealthy"
			break
		}
	}

	statusCode := fiber.StatusOK
	if overall == "unhealthy" {
		statusCode = fiber.StatusServiceUnavailable
	}

	resp := &HealthResponse{
		Status: overall,
		Checks: checks,
	}

	hc.Log.Infof("health check: status=%s", overall)

	return ctx.Status(statusCode).JSON(model.SuccessResponse(resp, "health check"))
}

func errorToString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func (hc *HealthController) Healthz(ctx fiber.Ctx) error {
	hc.Log.Debug("liveness probe hit")
	return ctx.SendStatus(fiber.StatusOK)
}

func (hc *HealthController) Readyz(ctx fiber.Ctx) error {
	hc.Log.Debug("readiness probe hit")
	return hc.Check(ctx)
}
