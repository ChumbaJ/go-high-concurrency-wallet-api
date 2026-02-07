// Package service
package service

import (
	"context"
	"fmt"

	"test-psql/internal/queue"
	"test-psql/pkg/logger"
)

type walletRepo interface {
	GetBalance(ctx context.Context, walletID string) (int64, error)
	Deposit(ctx context.Context, walletID string, amount int64) error
	Withdraw(ctx context.Context, walletID string, amount int64) error
}

type WalletService struct {
	queue *queue.Queue
	repo  walletRepo
}

func NewWalletService(queue *queue.Queue, repo walletRepo) *WalletService {
	return &WalletService{queue: queue, repo: repo}
}

func (s *WalletService) UpdateBalance(ctx context.Context, walletID string, operationType string, amount int64) error {
	logger.Info(fmt.Sprintf("service UpdateBalance walletId=%s op=%s amount=%d", walletID, operationType, amount))

	resultChan := make(chan error)

	switch operationType {
	case "DEPOSIT":
		s.queue.Add(ctx, "DEPOSIT", walletID, amount, resultChan)
		return <-resultChan
	case "WITHDRAW":
		s.queue.Add(ctx, "WITHDRAW", walletID, amount, resultChan)
		return <-resultChan
	default:
		return fmt.Errorf("unknown operation type: %s", operationType)
	}
}

func (s *WalletService) GetBalance(ctx context.Context, walletID string) (int64, error) {
	logger.Info(fmt.Sprintf("service GetBalance walletId=%s", walletID))
	return s.repo.GetBalance(ctx, walletID)
}
