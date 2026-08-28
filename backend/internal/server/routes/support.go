package routes

import (
	"github.com/Wei-Shaw/sub2api/internal/handler"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterSupportRoutes(v1 *gin.RouterGroup, h *handler.Handlers, panelRateLimiter *middleware.PanelRateLimiter) {
	support := v1.Group("/support")
	support.Use(panelRateLimiter.PublicIP())
	support.POST("/contact", h.SupportContact.Contact)
}
