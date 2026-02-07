// Package wallet dto schema
package dto

import "fmt"

type OperationType string

const (
	DEPOSIT  OperationType = "DEPOSIT"
	WITHDRAW OperationType = "WITHDRAW"
)

type UpdateWalletBalanceRequest struct {
	WalletID      string        `json:"walletId"`
	OperationType OperationType `json:"operationType"`
	Amount        int64         `json:"amount"`
}

func (r *UpdateWalletBalanceRequest) Validate() error {
	if r.WalletID == "" {
		return fmt.Errorf("walletId is required")
	}
	if r.OperationType != DEPOSIT && r.OperationType != WITHDRAW {
		return fmt.Errorf("operationType must be DEPOSIT or WITHDRAW")
	}
	if r.Amount <= 0 {
		return fmt.Errorf("amount must be positive")
	}
	return nil
}

type UpdateWalletBalanceResponse struct {
	WalletID string `json:"walletId"`
	Balance  int64  `json:"balance"`
}

type GetWalletBalanceResponse struct {
	WalletID string `json:"walletId"`
	Balance  int64  `json:"balance"`
}

type GetWalletBalanceRequest struct {
	WalletID string
}

func (r *GetWalletBalanceRequest) Validate() error {
	if r.WalletID == "" {
		return fmt.Errorf("walletId is required")
	}
	return nil
}
