package admin

import (
	"strconv"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type OperationalHandler struct {
	repo service.AdminOperationalRepository
}

func NewOperationalHandler(repo service.AdminOperationalRepository) *OperationalHandler {
	return &OperationalHandler{repo: repo}
}

func (h *OperationalHandler) GetModelMargins(c *gin.Context) {
	start, end := parseTimeRange(c)
	rows, err := h.repo.GetModelMargins(c.Request.Context(), start, end)
	if err != nil {
		response.InternalError(c, "Failed to get model margins")
		return
	}
	response.Success(c, gin.H{"models": rows, "start_date": start, "end_date": end})
}

func (h *OperationalHandler) GetUserModelBalances(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Query("user_id"), 10, 64)
	if err != nil || userID <= 0 {
		response.BadRequest(c, "Invalid user_id")
		return
	}
	rows, err := h.repo.GetUserModelBalances(c.Request.Context(), userID)
	if err != nil {
		response.InternalError(c, "Failed to get user model balances")
		return
	}
	response.Success(c, gin.H{"user_id": userID, "balances": rows})
}

func (h *OperationalHandler) GetSupplierPricing(c *gin.Context) {
	model := c.Query("model")
	if model == "" {
		response.BadRequest(c, "model is required")
		return
	}
	rows, err := h.repo.GetSupplierPricing(c.Request.Context(), model, time.Now())
	if err != nil {
		response.InternalError(c, "Failed to get supplier pricing")
		return
	}
	response.Success(c, gin.H{"model": model, "pricing": rows})
}

func (h *OperationalHandler) GetReconciliationDrift(c *gin.Context) {
	start, end := parseTimeRange(c)
	result, err := h.repo.GetReconciliationDrift(c.Request.Context(), start, end)
	if err != nil {
		response.InternalError(c, "Failed to get reconciliation drift")
		return
	}
	response.Success(c, gin.H{"drift": result, "start_date": start, "end_date": end})
}
