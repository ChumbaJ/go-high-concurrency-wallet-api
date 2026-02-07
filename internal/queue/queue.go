// Package queue for batching requests to db
package queue

import (
	"context"
	"time"
)

type opRequest struct {
	Op       string
	WalletID string
	Amount   int64
	Result   chan error
}

type walletRepo interface {
	GetBalance(ctx context.Context, walletID string) (int64, error)
	Deposit(ctx context.Context, walletID string, amount int64) error
	Withdraw(ctx context.Context, walletID string, amount int64) error
}

type Queue struct {
	opsChan     chan *opRequest
	walletRepo  walletRepo
	buffSize    int
	flushPeriod time.Duration
}

func NewQueue(walletRepo walletRepo, buffSize int, flushPeriod time.Duration) *Queue {
	return &Queue{
		opsChan:     make(chan *opRequest, buffSize),
		walletRepo:  walletRepo,
		buffSize:    buffSize,
		flushPeriod: flushPeriod,
	}
}

func (q *Queue) Add(ctx context.Context, op, walletID string, amount int64, result chan error) {
	q.opsChan <- &opRequest{Op: op, WalletID: walletID, Amount: amount, Result: result}
}

func (q *Queue) ProcessQueue(ctx context.Context) {
	ticker := time.NewTicker(q.flushPeriod)
	defer ticker.Stop()
	buff := make([]*opRequest, 0, q.buffSize)

	flush := func() {
		if len(buff) == 0 {
			return
		}
		batch := make([]*opRequest, len(buff))
		copy(batch, buff)
		go q.worker(ctx, batch)
		buff = buff[:0]
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			flush()
		case req := <-q.opsChan:
			buff = append(buff, req)
			if len(buff) >= q.buffSize {
				flush()
			}
		}
	}
}

func (q *Queue) worker(ctx context.Context, batch []*opRequest) {
	if len(batch) == 0 {
		return
	}
	type key struct {
		walletID string
		op       string
	}
	byKey := make(map[key][]*opRequest)
	for _, req := range batch {
		k := key{walletID: req.WalletID, op: req.Op}
		byKey[k] = append(byKey[k], req)
	}
	for k, requests := range byKey {
		var totalAmount int64
		for _, req := range requests {
			totalAmount += req.Amount
		}
		var err error
		switch k.op {
		case "DEPOSIT":
			err = q.walletRepo.Deposit(ctx, k.walletID, totalAmount)
		case "WITHDRAW":
			err = q.walletRepo.Withdraw(ctx, k.walletID, totalAmount)
		}
		for _, req := range requests {
			select {
			case req.Result <- err:
			default:
			}
		}
	}
}
