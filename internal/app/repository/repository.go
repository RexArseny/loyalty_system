package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

var (
	ErrNoOrders         = errors.New("no orders")
	ErrNotEnoughBalance = errors.New("not enough balance")
	ErrNoWithdrawals    = errors.New("no withdrawals")
)

type Repository interface {
	Registration(
		ctx context.Context,
		login string,
		hash string,
		salt string,
		userID uuid.UUID,
	) error
	GetUser(
		ctx context.Context,
		login string,
	) (*User, error)
	AddOrder(
		ctx context.Context,
		orderNumber string,
		userID uuid.UUID,
	) error
	GetOrders(
		ctx context.Context,
		userID uuid.UUID,
	) ([]Order, error)
	GetBalance(
		ctx context.Context,
		userID uuid.UUID,
	) (*Balance, error)
	Withdraw(
		ctx context.Context,
		orderNumber string,
		sum float64,
		userID uuid.UUID,
	) error
	GetWithdrawals(
		ctx context.Context,
		userID uuid.UUID,
	) ([]Withdraw, error)
	GetOrdersForUpdate(
		ctx context.Context,
	) ([]Order, error)
	UpdateOrder(
		ctx context.Context,
		orderNumber string,
		status string,
		accrual *float64,
		userID uuid.UUID,
	) error
	Close()
}

func NewRepository(
	ctx context.Context,
	logger *zap.Logger,
	databaseURI string,
) (Repository, error) {
	dbRepository, err := NewDBRepository(ctx, logger, databaseURI)
	if err != nil {
		return nil, fmt.Errorf("can not init db repository: %w", err)
	}
	return dbRepository, nil
}
