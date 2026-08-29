package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/payment"
)

const (
	gomerchDefaultAPIBase = "https://v1-api.pay-gomerch.web.id"
	gomerchHTTPTimeout    = 10 * time.Second
)

var ErrRefundUnsupported = errors.New("gomerch refunds are not supported")

type Gomerch struct {
	instanceID string
	config     map[string]string
	httpClient *http.Client
}

func NewGomerch(instanceID string, config map[string]string) (*Gomerch, error) {
	if strings.TrimSpace(config["apiKey"]) == "" {
		return nil, fmt.Errorf("gomerch config missing required key: apiKey")
	}

	cfg := cloneStringMap(config)
	if strings.TrimSpace(cfg["apiBase"]) == "" {
		cfg["apiBase"] = gomerchDefaultAPIBase
	}
	cfg["apiBase"] = strings.TrimRight(strings.TrimSpace(cfg["apiBase"]), "/")
	if strings.TrimSpace(cfg["currency"]) == "" {
		cfg["currency"] = "IDR"
	} else {
		currency, err := payment.NormalizePaymentCurrency(cfg["currency"])
		if err != nil {
			return nil, fmt.Errorf("gomerch config currency: %w", err)
		}
		cfg["currency"] = currency
	}
	if strings.TrimSpace(cfg["idrRate"]) == "" {
		cfg["idrRate"] = "16000"
	}

	return &Gomerch{
		instanceID: instanceID,
		config:     cfg,
		httpClient: &http.Client{Timeout: gomerchHTTPTimeout},
	}, nil
}

func (g *Gomerch) Name() string                          { return "GoMerch QRIS" }
func (g *Gomerch) ProviderKey() string                   { return payment.TypeGomerch }
func (g *Gomerch) SupportedTypes() []payment.PaymentType { return []payment.PaymentType{"qris"} }
func (g *Gomerch) MerchantIdentityMetadata() map[string]string {
	if g == nil {
		return nil
	}
	return map[string]string{"currency": g.config["currency"]}
}

func (g *Gomerch) request(ctx context.Context, path string, payload any) ([]byte, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, g.config["apiBase"]+path, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-API-Key", g.config["apiKey"])
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("gomerch: upstream status %d: %s", resp.StatusCode, strings.TrimSpace(string(data)))
	}
	return data, nil
}

func (g *Gomerch) CreatePayment(ctx context.Context, req payment.CreatePaymentRequest) (*payment.CreatePaymentResponse, error) {
	amount, err := payment.AmountToMinorUnit(req.Amount, "IDR")
	if err != nil {
		return nil, fmt.Errorf("gomerch create payment: %w", err)
	}

	var resp struct {
		Success    bool   `json:"success"`
		PaymentRef string `json:"payment_ref"`
		QRString   string `json:"qr_string"`
	}
	data, err := g.request(ctx, "/api/qris/create", struct {
		Amount   int64 `json:"amount"`
		StaticQR bool  `json:"static_qr"`
	}{Amount: amount, StaticQR: false})
	if err != nil {
		return nil, fmt.Errorf("gomerch create payment: %w", err)
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("gomerch create payment response: %w", err)
	}
	if !resp.Success || strings.TrimSpace(resp.PaymentRef) == "" {
		return nil, fmt.Errorf("gomerch create payment: response missing successful payment_ref")
	}
	return &payment.CreatePaymentResponse{
		TradeNo:    resp.PaymentRef,
		QRCode:     resp.QRString,
		Currency:   "IDR",
		ResultType: payment.CreatePaymentResultOrderCreated,
	}, nil
}

func (g *Gomerch) QueryOrder(ctx context.Context, tradeNo string) (*payment.QueryOrderResponse, error) {
	var resp struct {
		Success    bool   `json:"success"`
		PaymentRef string `json:"payment_ref"`
		Status     string `json:"status"`
		Amount     int64  `json:"amount"`
	}
	data, err := g.request(ctx, "/api/qris/status", map[string]string{"payment_ref": tradeNo})
	if err != nil {
		return nil, fmt.Errorf("gomerch query order: %w", err)
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("gomerch query order response: %w", err)
	}
	if !resp.Success || strings.TrimSpace(resp.PaymentRef) == "" {
		return nil, fmt.Errorf("gomerch query order: response missing successful payment_ref")
	}

	status := payment.ProviderStatusPending
	switch strings.ToUpper(resp.Status) {
	case "PAID":
		status = payment.ProviderStatusPaid
	case "EXPIRED":
		status = payment.ProviderStatusFailed
	case "PENDING", "AMBIGUOUS":
		status = payment.ProviderStatusPending
	}
	return &payment.QueryOrderResponse{
		TradeNo: resp.PaymentRef,
		Amount:  float64(resp.Amount),
		Status:  status,
	}, nil
}

func (g *Gomerch) VerifyNotification(context.Context, string, map[string]string) (*payment.PaymentNotification, error) {
	return nil, errors.New("gomerch does not support webhooks")
}

func (g *Gomerch) Refund(context.Context, payment.RefundRequest) (*payment.RefundResponse, error) {
	return nil, ErrRefundUnsupported
}

func (g *Gomerch) CancelPayment(context.Context, string) error { return nil }

var _ payment.Provider = (*Gomerch)(nil)
var _ payment.CancelableProvider = (*Gomerch)(nil)
var _ payment.MerchantIdentityProvider = (*Gomerch)(nil)
