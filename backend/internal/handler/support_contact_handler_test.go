package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type supportEmailSenderMock struct {
	calls int
	err   error
}

func (m *supportEmailSenderMock) SendEmail(context.Context, string, string, string) error {
	m.calls++
	return m.err
}

func supportContactTestRouter(sender supportEmailSenderMock) (*gin.Engine, *supportEmailSenderMock) {
	gin.SetMode(gin.TestMode)
	mock := &sender
	h := &SupportContactHandler{emailSender: mock, now: time.Now, buckets: make(map[string]supportContactBucket)}
	r := gin.New()
	r.POST("/contact", h.Contact)
	return r, mock
}

func performSupportContactRequest(r *gin.Engine, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/contact", jsonReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func jsonReader(body string) *strings.Reader { return strings.NewReader(body) }

func TestSupportContactHandlerSuccess(t *testing.T) {
	r, mock := supportContactTestRouter(supportEmailSenderMock{})
	w := performSupportContactRequest(r, `{"name":"Ada","email":"ada@example.com","subject":"Help","message":"Please help me with this issue."}`)
	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, 1, mock.calls)
}

func TestSupportContactHandlerValidation(t *testing.T) {
	r, mock := supportContactTestRouter(supportEmailSenderMock{})
	w := performSupportContactRequest(r, `{"name":"","email":"bad","subject":"","message":"short"}`)
	require.Equal(t, http.StatusBadRequest, w.Code)
	require.Equal(t, 0, mock.calls)
}

func TestSupportContactHandlerRateLimit(t *testing.T) {
	r, mock := supportContactTestRouter(supportEmailSenderMock{})
	body := `{"name":"Ada","email":"ada@example.com","subject":"Help","message":"Please help me with this issue."}`
	for i := 0; i < supportContactLimit; i++ {
		require.Equal(t, http.StatusOK, performSupportContactRequest(r, body).Code)
	}
	require.Equal(t, http.StatusTooManyRequests, performSupportContactRequest(r, body).Code)
	require.Equal(t, supportContactLimit, mock.calls)
}

func TestSupportContactHandlerEmailFailure(t *testing.T) {
	r, _ := supportContactTestRouter(supportEmailSenderMock{err: errors.New("smtp unavailable")})
	w := performSupportContactRequest(r, `{"name":"Ada","email":"ada@example.com","subject":"Help","message":"Please help me with this issue."}`)
	require.Equal(t, http.StatusServiceUnavailable, w.Code)
}
