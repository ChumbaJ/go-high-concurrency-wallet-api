package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"test-psql/internal/queue"
)

type stubWalletRepo struct {
	getBalanceVal int64
	getBalanceErr error
	depositErr    error
	withdrawErr   error
}

func (s *stubWalletRepo) GetBalance(ctx context.Context, walletID string) (int64, error) {
	return s.getBalanceVal, s.getBalanceErr
}

func (s *stubWalletRepo) Deposit(ctx context.Context, walletID string, amount int64) error {
	return s.depositErr
}

func (s *stubWalletRepo) Withdraw(ctx context.Context, walletID string, amount int64) error {
	return s.withdrawErr
}

func TestWalletService_UpdateBalance(t *testing.T) {
	t.Run("DEPOSIT ok", func(t *testing.T) {
		repo := &stubWalletRepo{}
		q := queue.NewQueue(repo, 50, 100*time.Millisecond)
		go q.ProcessQueue(context.Background())
		svc := NewWalletService(q, repo)
		err := svc.UpdateBalance(context.Background(), "id1", "DEPOSIT", 100)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("DEPOSIT repo error", func(t *testing.T) {
		repo := &stubWalletRepo{depositErr: errors.New("db error")}
		q := queue.NewQueue(repo, 50, 100*time.Millisecond)
		go q.ProcessQueue(context.Background())
		svc := NewWalletService(q, repo)
		err := svc.UpdateBalance(context.Background(), "id1", "DEPOSIT", 100)
		if err == nil || err.Error() != "db error" {
			t.Errorf("want db error, got %v", err)
		}
	})

	t.Run("WITHDRAW ok", func(t *testing.T) {
		repo := &stubWalletRepo{}
		q := queue.NewQueue(repo, 50, 100*time.Millisecond)
		go q.ProcessQueue(context.Background())
		svc := NewWalletService(q, repo)
		err := svc.UpdateBalance(context.Background(), "id1", "WITHDRAW", 50)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("WITHDRAW repo error", func(t *testing.T) {
		repo := &stubWalletRepo{withdrawErr: errors.New("insufficient balance")}
		q := queue.NewQueue(repo, 50, 100*time.Millisecond)
		go q.ProcessQueue(context.Background())
		svc := NewWalletService(q, repo)
		err := svc.UpdateBalance(context.Background(), "id1", "WITHDRAW", 50)
		if err == nil || err.Error() != "insufficient balance" {
			t.Errorf("want insufficient balance, got %v", err)
		}
	})

	t.Run("unknown operation", func(t *testing.T) {
		repo := &stubWalletRepo{}
		q := queue.NewQueue(repo, 50, 100*time.Millisecond)
		go q.ProcessQueue(context.Background())
		svc := NewWalletService(q, repo)
		err := svc.UpdateBalance(context.Background(), "id1", "UNKNOWN", 10)
		if err == nil {
			t.Fatal("expected error for unknown operation")
		}
		if err.Error() != "unknown operation type: UNKNOWN" {
			t.Errorf("got %v", err)
		}
	})
}

func TestWalletService_GetBalance(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		repo := &stubWalletRepo{getBalanceVal: 999}
		q := queue.NewQueue(repo, 50, 100*time.Millisecond)
		go q.ProcessQueue(context.Background())
		svc := NewWalletService(q, repo)
		balance, err := svc.GetBalance(context.Background(), "id1")
		if err != nil {
			t.Fatal(err)
		}
		if balance != 999 {
			t.Errorf("got balance %d, want 999", balance)
		}
	})

	t.Run("repo error", func(t *testing.T) {
		repo := &stubWalletRepo{getBalanceErr: errors.New("not found")}
		q := queue.NewQueue(repo, 50, 100*time.Millisecond)
		go q.ProcessQueue(context.Background())
		svc := NewWalletService(q, repo)
		_, err := svc.GetBalance(context.Background(), "id1")
		if err == nil || err.Error() != "not found" {
			t.Errorf("want not found, got %v", err)
		}
	})
}
