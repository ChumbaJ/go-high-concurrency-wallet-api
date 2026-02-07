// Package handlers
package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"test-psql/internal/http/dto"
	"test-psql/pkg/logger"
	"time"
)

type walletService interface {
	UpdateBalance(ctx context.Context, walletID string, operationType string, amount int64) error
	GetBalance(ctx context.Context, walletID string) (int64, error)
}

type WalletHandler struct {
	service      walletService
	requestTimeout time.Duration
}

func NewWalletHandler(svc walletService, timeout time.Duration) *WalletHandler {
	return &WalletHandler{
		service:       svc,
		requestTimeout: timeout,
	}
}

func (h *WalletHandler) UpdateWalletBalance(w http.ResponseWriter, r *http.Request) {
	logger.Info("POST /api/v1/wallet")
	if r.Method != http.MethodPost {
		logger.Error(fmt.Sprintf("method not allowed: %s", r.Method))
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), h.requestTimeout)
	defer cancel()

	var req dto.UpdateWalletBalanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		var typeErr *json.UnmarshalTypeError
		if errors.As(err, &typeErr) && typeErr.Field == "amount" {
			logger.Error(fmt.Sprintf("invalid amount type: %v", err))
			http.Error(w, "amount must be a number", http.StatusBadRequest)
			return
		}
		logger.Error(fmt.Sprintf("invalid request body: %v", err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := req.Validate(); err != nil {
		logger.Error(fmt.Sprintf("validation error: %v", err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.service.UpdateBalance(ctx, req.WalletID, string(req.OperationType), req.Amount); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			logger.Error("request timeout")
			http.Error(w, "Request timeout", http.StatusRequestTimeout)
			return
		}
		logger.Error(fmt.Sprintf("update balance failed: %v", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	balance, err := h.service.GetBalance(ctx, req.WalletID)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			logger.Error("request timeout")
			http.Error(w, "Request timeout", http.StatusRequestTimeout)
			return
		}
		logger.Error(fmt.Sprintf("get balance failed: %v", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := dto.UpdateWalletBalanceResponse{
		WalletID: req.WalletID,
		Balance:  balance,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
	logger.Info(fmt.Sprintf("balance updated: walletId=%s balance=%d", req.WalletID, balance))
}

func (h *WalletHandler) GetWalletBalance(w http.ResponseWriter, r *http.Request) {
	logger.Info("GET /api/v1/wallets/{id}")
	if r.Method != http.MethodGet {
		logger.Error(fmt.Sprintf("method not allowed: %s", r.Method))
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), h.requestTimeout)
	defer cancel()

	path := strings.TrimPrefix(r.URL.Path, "/api/v1/wallets/")
	walletID := strings.TrimSuffix(path, "/")

	req := dto.GetWalletBalanceRequest{WalletID: walletID}
	if err := req.Validate(); err != nil {
		logger.Error(fmt.Sprintf("validation error: %v", err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	balance, err := h.service.GetBalance(ctx, req.WalletID)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			logger.Error("request timeout")
			http.Error(w, "Request timeout", http.StatusRequestTimeout)
			return
		}
		logger.Error(fmt.Sprintf("get balance failed: %v", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := dto.GetWalletBalanceResponse{
		WalletID: req.WalletID,
		Balance:  balance,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
	logger.Info(fmt.Sprintf("balance retrieved: walletId=%s balance=%d", req.WalletID, balance))
}
