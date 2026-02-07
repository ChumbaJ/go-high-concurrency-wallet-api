package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type mockWalletService struct {
	updateBalanceErr error
	getBalanceVal    int64
	getBalanceErr    error
}

func (m *mockWalletService) UpdateBalance(ctx context.Context, walletID, operationType string, amount int64) error {
	return m.updateBalanceErr
}

func (m *mockWalletService) GetBalance(ctx context.Context, walletID string) (int64, error) {
	return m.getBalanceVal, m.getBalanceErr
}

func TestWalletHandler_UpdateWalletBalance(t *testing.T) {
	validReqBody := map[string]any{
		"walletId":      "550e8400-e29b-41d4-a716-446655440000",
		"operationType": "DEPOSIT",
		"amount":        1000,
	}
	validBody, _ := json.Marshal(validReqBody)

	t.Run("ok", func(t *testing.T) {
		svc := &mockWalletService{getBalanceVal: 1500}
		h := NewWalletHandler(svc, 30*time.Second)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/wallet", bytes.NewReader(validBody))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		h.UpdateWalletBalance(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("got status %d, want 200", rec.Code)
		}
		var res struct {
			WalletID string `json:"walletId"`
			Balance  int64  `json:"balance"`
		}
		if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
			t.Fatal(err)
		}
		if res.Balance != 1500 {
			t.Errorf("got balance %d, want 1500", res.Balance)
		}
	})

	t.Run("invalid body", func(t *testing.T) {
		h := NewWalletHandler(&mockWalletService{}, 30*time.Second)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/wallet", bytes.NewReader([]byte("{")))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		h.UpdateWalletBalance(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("got status %d, want 400", rec.Code)
		}
	})

	t.Run("service error", func(t *testing.T) {
		svc := &mockWalletService{updateBalanceErr: errors.New("service error")}
		h := NewWalletHandler(svc, 30*time.Second)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/wallet", bytes.NewReader(validBody))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		h.UpdateWalletBalance(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Errorf("got status %d, want 500", rec.Code)
		}
	})

	t.Run("method not allowed", func(t *testing.T) {
		h := NewWalletHandler(&mockWalletService{}, 30*time.Second)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/wallet", nil)
		rec := httptest.NewRecorder()

		h.UpdateWalletBalance(rec, req)

		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("got status %d, want 405", rec.Code)
		}
	})

	t.Run("amount is string", func(t *testing.T) {
		h := NewWalletHandler(&mockWalletService{}, 30*time.Second)
		body := []byte(`{"walletId":"550e8400-e29b-41d4-a716-446655440000","operationType":"DEPOSIT","amount":"1000"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/wallet", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		h.UpdateWalletBalance(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("got status %d, want 400", rec.Code)
		}
		if got := rec.Body.String(); got != "amount must be a number\n" {
			t.Fatalf("unexpected body: %q", got)
		}
	})
}

func TestWalletHandler_GetWalletBalance(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		svc := &mockWalletService{getBalanceVal: 2000}
		h := NewWalletHandler(svc, 30*time.Second)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/wallets/550e8400-e29b-41d4-a716-446655440000", nil)
		rec := httptest.NewRecorder()

		h.GetWalletBalance(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("got status %d, want 200", rec.Code)
		}
		var res struct {
			WalletID string `json:"walletId"`
			Balance  int64  `json:"balance"`
		}
		if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
			t.Fatal(err)
		}
		if res.Balance != 2000 {
			t.Errorf("got balance %d, want 2000", res.Balance)
		}
	})

	t.Run("empty wallet id", func(t *testing.T) {
		h := NewWalletHandler(&mockWalletService{}, 30*time.Second)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/wallets/", nil)
		rec := httptest.NewRecorder()

		h.GetWalletBalance(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("got status %d, want 400", rec.Code)
		}
	})

	t.Run("service error", func(t *testing.T) {
		svc := &mockWalletService{getBalanceErr: errors.New("service error")}
		h := NewWalletHandler(svc, 30*time.Second)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/wallets/550e8400-e29b-41d4-a716-446655440000", nil)
		rec := httptest.NewRecorder()

		h.GetWalletBalance(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Errorf("got status %d, want 500", rec.Code)
		}
	})
}
