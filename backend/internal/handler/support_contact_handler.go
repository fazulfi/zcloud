package handler

import (
	"context"
	"net/http"
	"net/mail"
	"os"
	"strings"
	"sync"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

const (
	supportEmailFallback = "support@zrouter.dev"
	supportContactLimit  = 5
	supportContactWindow = time.Hour
)

type supportEmailSender interface {
	SendEmail(ctx context.Context, to, subject, body string) error
}

type supportContactRequest struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Subject string `json:"subject"`
	Message string `json:"message"`
}

type supportContactBucket struct {
	startedAt time.Time
	count     int
}

// SupportContactHandler handles unauthenticated support contact requests.
type SupportContactHandler struct {
	emailSender supportEmailSender
	now         func() time.Time
	mu          sync.Mutex
	buckets     map[string]supportContactBucket
}

func NewSupportContactHandler(emailService *service.EmailService) *SupportContactHandler {
	return &SupportContactHandler{
		emailSender: emailService,
		now:         time.Now,
		buckets:     make(map[string]supportContactBucket),
	}
}

func (h *SupportContactHandler) Contact(c *gin.Context) {
	ip := middleware.SecurityClientIP(c)
	if !h.allow(ip) {
		response.ErrorWithDetails(c, http.StatusTooManyRequests, "Too many requests, please try again later", "SUPPORT_CONTACT_RATE_LIMITED", nil)
		return
	}

	var req supportContactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "request body must be valid JSON")
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	req.Email = strings.TrimSpace(req.Email)
	req.Subject = strings.TrimSpace(req.Subject)
	req.Message = strings.TrimSpace(req.Message)
	if req.Name == "" {
		response.BadRequest(c, "name is required")
		return
	}
	parsedEmail, err := mail.ParseAddress(req.Email)
	if err != nil || parsedEmail.Address != req.Email {
		response.BadRequest(c, "email must be a valid email address")
		return
	}
	if len([]rune(req.Subject)) < 1 || len([]rune(req.Subject)) > 200 {
		response.BadRequest(c, "subject must be between 1 and 200 characters")
		return
	}
	if length := len([]rune(req.Message)); length < 10 || length > 5000 {
		response.BadRequest(c, "message must be between 10 and 5000 characters")
		return
	}
	if h.emailSender == nil {
		response.ErrorFrom(c, infraerrors.ServiceUnavailable("SUPPORT_EMAIL_UNAVAILABLE", "support email service is unavailable"))
		return
	}

	target := strings.TrimSpace(os.Getenv("SUPPORT_EMAIL"))
	if target == "" {
		target = supportEmailFallback
	}
	body := "Name: " + req.Name + "\nEmail: " + req.Email + "\n\n" + req.Message
	if err := h.emailSender.SendEmail(c.Request.Context(), target, req.Subject, body); err != nil {
		response.ErrorFrom(c, infraerrors.ServiceUnavailable("SUPPORT_EMAIL_UNAVAILABLE", "support email service is unavailable"))
		return
	}

	response.Success(c, gin.H{"message": "Your support request has been sent"})
}

func (h *SupportContactHandler) allow(ip string) bool {
	now := h.now()
	h.mu.Lock()
	defer h.mu.Unlock()
	bucket := h.buckets[ip]
	if bucket.startedAt.IsZero() || now.Sub(bucket.startedAt) >= supportContactWindow {
		bucket = supportContactBucket{startedAt: now}
	}
	if bucket.count >= supportContactLimit {
		h.buckets[ip] = bucket
		return false
	}
	bucket.count++
	h.buckets[ip] = bucket
	return true
}
