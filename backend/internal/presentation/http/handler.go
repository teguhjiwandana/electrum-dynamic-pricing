package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/electrum/dynamic-pricing-engine/internal/application"
	"github.com/electrum/dynamic-pricing-engine/internal/domain/pricing"
	"github.com/electrum/dynamic-pricing-engine/internal/infrastructure/auth"
	"github.com/gin-gonic/gin"
)

// Handler holds all HTTP handler dependencies.
type Handler struct {
	uc     *application.PricingUseCase
	jwtSvc *auth.JWTService
}

// NewHandler creates a new Handler.
func NewHandler(uc *application.PricingUseCase, jwtSvc *auth.JWTService) *Handler {
	return &Handler{uc: uc, jwtSvc: jwtSvc}
}

// Login authenticates a user and returns a JWT token.
func (h *Handler) Login(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, "username and password required")
		return
	}

	token, exp, username, role, err := h.jwtSvc.Authenticate(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		Error(c, http.StatusUnauthorized, "invalid credentials")
		return
	}

	Success(c, http.StatusOK, gin.H{
		"token":      token,
		"expires_at": exp,
		"username":   username,
		"role":       role,
	})
}

// GetPricing calculates a rental price.
func (h *Handler) GetPricing(c *gin.Context) {
	input, err := parsePricingParams(c)
	if err != nil {
		Error(c, http.StatusBadRequest, err.Error())
		return
	}

	result, err := h.uc.CalculatePrice(c.Request.Context(), input)
	if err != nil {
		Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	Success(c, http.StatusOK, result)
}

// GetBreakdown returns the detailed pricing breakdown.
func (h *Handler) GetBreakdown(c *gin.Context) {
	input, err := parsePricingParams(c)
	if err != nil {
		Error(c, http.StatusBadRequest, err.Error())
		return
	}

	result, err := h.uc.GetBreakdown(c.Request.Context(), input)
	if err != nil {
		Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	Success(c, http.StatusOK, result)
}

// GetConfig returns the current pricing configuration.
func (h *Handler) GetConfig(c *gin.Context) {
	cfg, err := h.uc.GetConfig(c.Request.Context())
	if err != nil {
		Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	Success(c, http.StatusOK, cfg)
}

// UpdateConfig updates the pricing configuration.
func (h *Handler) UpdateConfig(c *gin.Context) {
	var req struct {
		BasePricePerHour float64          `json:"base_price_per_hour"`
		Currency         string           `json:"currency"`
		SurgeCap         float64          `json:"surge_cap_multiplier"`
		DemandRules      json.RawMessage  `json:"demand_multipliers"`
		ZoneSurge        json.RawMessage  `json:"zone_surge_config"`
		BatteryDiscounts json.RawMessage  `json:"battery_discount_tiers"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, "invalid request body")
		return
	}

	changedBy, _ := c.Get("username")

	cfg, err := h.uc.UpdateConfig(
		c.Request.Context(),
		req.BasePricePerHour,
		req.Currency,
		req.SurgeCap,
		req.DemandRules,
		req.ZoneSurge,
		req.BatteryDiscounts,
		changedBy.(string),
	)
	if err != nil {
		Error(c, http.StatusBadRequest, err.Error())
		return
	}

	Success(c, http.StatusOK, cfg)
}

// GetConfigHistory returns config version history.
func (h *Handler) GetConfigHistory(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	result, err := h.uc.GetConfigHistory(c.Request.Context(), page, pageSize)
	if err != nil {
		Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	Paginated(c, result.Data, result.Total, result.Page, result.PageSize, result.TotalPages)
}

// GetAuditLogs returns paginated audit log entries.
func (h *Handler) GetAuditLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	vehicleID := c.Query("vehicle_id")
	zone := c.Query("zone")

	result, err := h.uc.GetAuditLogs(c.Request.Context(), page, pageSize, vehicleID, zone)
	if err != nil {
		Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	Paginated(c, result.Data, result.Total, result.Page, result.PageSize, result.TotalPages)
}

// GetZones returns zone utilization data.
func (h *Handler) GetZones(c *gin.Context) {
	zones, err := h.uc.ListZones(c.Request.Context())
	if err != nil {
		Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	Success(c, http.StatusOK, zones)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func parsePricingParams(c *gin.Context) (pricing.PricingInput, error) {
	vehicleID := c.Query("vehicle_id")
	zone := c.Query("zone")
	durStr := c.Query("duration_hours")

	if vehicleID == "" || zone == "" || durStr == "" {
		return pricing.PricingInput{}, fmt.Errorf("vehicle_id, zone, and duration_hours are required")
	}

	dur, err := strconv.Atoi(durStr)
	if err != nil || dur < 1 || dur > 720 {
		return pricing.PricingInput{}, fmt.Errorf("duration_hours must be between 1 and 720")
	}

	return pricing.PricingInput{
		VehicleID:     vehicleID,
		Zone:          zone,
		DurationHours: dur,
	}, nil
}
