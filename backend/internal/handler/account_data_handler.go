package handler

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/apikey"
	"github.com/Wei-Shaw/sub2api/ent/usersubscription"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type AccountDataHandler struct {
	userService         *service.UserService
	apiKeyService       *service.APIKeyService
	subscriptionService *service.SubscriptionService
	paymentService      *service.PaymentService
	entClient           *dbent.Client
}

func NewAccountDataHandler(userService *service.UserService, apiKeyService *service.APIKeyService, subscriptionService *service.SubscriptionService, paymentService *service.PaymentService, entClient *dbent.Client) *AccountDataHandler {
	return &AccountDataHandler{userService: userService, apiKeyService: apiKeyService, subscriptionService: subscriptionService, paymentService: paymentService, entClient: entClient}
}

func (h *AccountDataHandler) ExportInvoices(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	orders, _, err := h.paymentService.GetUserOrders(c.Request.Context(), subject.UserID, service.OrderListParams{Page: 1, PageSize: 10000})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	rows := make([]map[string]any, 0, len(orders))
	for _, o := range orders {
		rows = append(rows, map[string]any{"id": o.ID, "amount": o.Amount, "status": o.Status, "order_type": o.OrderType, "created_at": o.CreatedAt})
	}
	writeExport(c, "invoices", rows, []string{"id", "amount", "status", "order_type", "created_at"})
}

func (h *AccountDataHandler) DeleteAccount(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	ctx := c.Request.Context()
	keys, err := h.entClient.APIKey.Query().Where(apikey.UserIDEQ(subject.UserID), apikey.DeletedAtIsNil()).All(ctx)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	for _, key := range keys {
		if err := h.apiKeyService.Delete(ctx, key.ID, subject.UserID); err != nil {
			response.ErrorFrom(c, err)
			return
		}
	}
	subs, err := h.entClient.UserSubscription.Query().Where(usersubscription.UserIDEQ(subject.UserID), usersubscription.DeletedAtIsNil()).All(ctx)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	for _, sub := range subs {
		if err := h.subscriptionService.RevokeSubscription(ctx, sub.ID); err != nil {
			response.ErrorFrom(c, err)
			return
		}
	}
	if err := h.userService.Delete(ctx, subject.UserID); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	c.JSON(http.StatusAccepted, gin.H{"code": "deletion_scheduled", "message": "account deletion scheduled"})
}

func writeExport(c *gin.Context, name string, rows []map[string]any, columns []string) {
	format := strings.ToLower(c.DefaultQuery("format", "json"))
	if format != "json" && format != "csv" {
		response.BadRequest(c, "format must be csv or json")
		return
	}
	if format == "json" {
		b, _ := json.Marshal(rows)
		c.Data(http.StatusOK, "application/json; charset=utf-8", b)
		return
	}
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s.csv", name))
	c.Header("Content-Type", "text/csv; charset=utf-8")
	w := csv.NewWriter(c.Writer)
	_ = w.Write(columns)
	for _, row := range rows {
		values := make([]string, len(columns))
		for i, column := range columns {
			values[i] = fmt.Sprint(row[column])
		}
		_ = w.Write(values)
	}
	w.Flush()
}

func exportTime(v string) time.Time { t, _ := time.Parse(time.RFC3339, v); return t }
