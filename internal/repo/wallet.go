// Package repo
package repo

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"test-psql/internal/models"
	"test-psql/pkg/logger"
)

type WalletRepo struct {
	db *gorm.DB
}

func NewWalletRepo(db *gorm.DB) *WalletRepo {
	return &WalletRepo{db: db}
}

func (r *WalletRepo) GetBalance(ctx context.Context, walletID string) (int64, error) {
	logger.Info(fmt.Sprintf("repo GetBalance walletId=%s", walletID))
	var w models.Wallet
	err := r.db.WithContext(ctx).Select("balance").Where("id = ?", walletID).First(&w).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, nil
		}
		logger.Error(fmt.Sprintf("repo GetBalance db error: %v", err))
		return 0, err
	}
	return w.Balance, nil
}

func (r *WalletRepo) Deposit(ctx context.Context, walletID string, amount int64) error {
	logger.Info(fmt.Sprintf("repo Deposit walletId=%s amount=%d", walletID, amount))
	if amount <= 0 {
		return fmt.Errorf("amount must be positive")
	}
	result := r.db.WithContext(ctx).Model(&models.Wallet{}).
		Where("id = ?", walletID).
		Update("balance", gorm.Expr("balance + ?", amount))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("wallet not found")
	}
	return nil
}

func (r *WalletRepo) Withdraw(ctx context.Context, walletID string, amount int64) error {
	logger.Info(fmt.Sprintf("repo Withdraw walletId=%s amount=%d", walletID, amount))
	if amount <= 0 {
		return fmt.Errorf("amount must be positive")
	}
	result := r.db.WithContext(ctx).Model(&models.Wallet{}).
		Where("id = ? AND balance >= ?", walletID, amount).
		Update("balance", gorm.Expr("balance - ?", amount))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		var w models.Wallet
		err := r.db.WithContext(ctx).Where("id = ?", walletID).First(&w).Error
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("wallet not found")
		}
		return fmt.Errorf("insufficient balance")
	}
	return nil
}
