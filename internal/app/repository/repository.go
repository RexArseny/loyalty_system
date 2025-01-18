package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

var (
	ErrOriginalLoginUniqueViolation = errors.New("login is not unique")
	ErrInvalidAuthData              = errors.New("login or password is incorrect")
	ErrAlreadyAdded                 = errors.New("order has been already added")
	ErrAlreadyAddedByAnotherUser    = errors.New("order has been already added by another user")
	ErrInvalidOrderNumber           = errors.New("invalid order number")
	ErrNoOrders                     = errors.New("no orders")
	ErrNotEnoughBalance             = errors.New("not enough balance")
	ErrNoWithdrawals                = errors.New("no withdrawals")
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
		sum int,
		userID uuid.UUID,
	) error
	GetWithdrawals(
		ctx context.Context,
		userID uuid.UUID,
	) ([]Withdraw, error)
	GetOrderForUpdate(
		ctx context.Context,
	) (*string, *uuid.UUID, error)
	UpdateOrder(
		ctx context.Context,
		orderNumber string,
		status string,
		accrual *int,
		userID *uuid.UUID,
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
