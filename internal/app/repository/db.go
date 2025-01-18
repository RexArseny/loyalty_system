package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/RexArseny/loyalty_system/internal/app/external"
	"github.com/RexArseny/loyalty_system/internal/app/models"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"
)

type DBRepository struct {
	logger *zap.Logger
	pool   *Pool
}

func NewDBRepository(ctx context.Context, logger *zap.Logger, connString string) (*DBRepository, error) {
	m, err := migrate.New("file://./internal/app/repository/migrations", connString)
	if err != nil {
		return nil, fmt.Errorf("can not create migration instance: %w", err)
	}
	err = m.Up()
	if err != nil {
		if !errors.Is(err, migrate.ErrNoChange) {
			return nil, fmt.Errorf("can not migrate up: %w", err)
		}
	}

	pool, err := NewPool(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("can not create new pool: %w", err)
	}
	err = pool.Ping(ctx)
	if err != nil {
		return nil, fmt.Errorf("can not ping PostgreSQL server: %w", err)
	}

	return &DBRepository{
		logger: logger,
		pool:   pool,
	}, nil
}

func (d *DBRepository) Registration(
	ctx context.Context,
	login string,
	hash string,
	salt string,
	userID uuid.UUID,
) error {
	tx, err := d.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("can not start transaction: %w", err)
	}
	defer func() {
		err = tx.Rollback(ctx)
		if err != nil && !strings.Contains(err.Error(), "tx is closed") {
			d.logger.Error("Can not rollback transaction", zap.Error(err))
		}
	}()

	_, err = tx.Exec(ctx, `INSERT INTO users (user_id, login, hash, salt)
							VALUES ($1, $2, $3, $4)`, userID, login, hash, salt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return ErrOriginalLoginUniqueViolation
		}
		return fmt.Errorf("can not add user: %w", err)
	}

	_, err = tx.Exec(ctx, `INSERT INTO balances (user_id, balance, withdrawn)
							VALUES ($1, $2, $3)`, userID, 0, 0)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return ErrOriginalLoginUniqueViolation
		}
		return fmt.Errorf("can not add user: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("can not commit transaction: %w", err)
	}

	return nil
}

func (d *DBRepository) GetUser(ctx context.Context, login string) (*User, error) {
	var user User
	err := d.pool.QueryRow(ctx, `SELECT user_id, login, hash, salt
								FROM users
								WHERE login = $1`, login).Scan(&user.UserID, &user.Login, &user.Hash, &user.Salt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrInvalidAuthData
		}
		return nil, fmt.Errorf("can not get user: %w", err)
	}

	return &user, nil
}

func (d *DBRepository) AddOrder(ctx context.Context, orderNumber string, userID uuid.UUID) error {
	tx, err := d.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("can not start transaction: %w", err)
	}
	defer func() {
		err = tx.Rollback(ctx)
		if err != nil && !strings.Contains(err.Error(), "tx is closed") {
			d.logger.Error("Can not rollback transaction", zap.Error(err))
		}
	}()

	var orderUserID uuid.UUID
	err = tx.QueryRow(ctx, "SELECT user_id FROM orders WHERE order_id = $1", orderNumber).Scan(&orderUserID)
	if err == nil {
		if orderUserID == userID {
			return ErrAlreadyAdded
		}
		return ErrAlreadyAddedByAnotherUser
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("can not get order: %w", err)
	}

	_, err = tx.Exec(ctx, `INSERT INTO orders (order_id, status, uploaded_at, user_id)
							VALUES ($1, $2, $3, $4)`,
		orderNumber,
		models.StatusNew,
		time.Now(),
		userID)
	if err != nil {
		return fmt.Errorf("can not add order: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("can not commit transaction: %w", err)
	}

	return nil
}

func (d *DBRepository) GetOrders(ctx context.Context, userID uuid.UUID) ([]Order, error) {
	rows, err := d.pool.Query(ctx, `SELECT order_id, status, accrual, uploaded_at 
									FROM orders 
									WHERE user_id = $1 
									ORDER BY uploaded_at`, userID)
	if err != nil {
		return nil, fmt.Errorf("can not get orders: %w", err)
	}
	defer rows.Close()

	var orders []Order
	for rows.Next() {
		var order Order
		err = rows.Scan(
			&order.Number,
			&order.Status,
			&order.Accrual,
			&order.UploadedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("can not read row: %w", err)
		}

		orders = append(orders, order)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("can not read rows: %w", rows.Err())
	}
	if len(orders) == 0 {
		return nil, ErrNoOrders
	}

	return orders, nil
}

func (d *DBRepository) GetBalance(ctx context.Context, userID uuid.UUID) (*Balance, error) {
	var balance Balance
	err := d.pool.QueryRow(ctx, `SELECT balance, withdrawn
								FROM balances
								WHERE user_id = $1`, userID).Scan(&balance.Current, &balance.Withdrawn)
	if err != nil {
		return nil, fmt.Errorf("can not get balance: %w", err)
	}

	return &balance, nil
}

func (d *DBRepository) Withdraw(ctx context.Context, orderNumber string, sum float64, userID uuid.UUID) error {
	tx, err := d.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("can not start transaction: %w", err)
	}
	defer func() {
		err = tx.Rollback(ctx)
		if err != nil && !strings.Contains(err.Error(), "tx is closed") {
			d.logger.Error("Can not rollback transaction", zap.Error(err))
		}
	}()

	var balance Balance
	err = tx.QueryRow(ctx, `SELECT balance, withdrawn
							FROM balances
							WHERE user_id = $1`, userID).Scan(&balance.Current, &balance.Withdrawn)
	if err != nil {
		return fmt.Errorf("can not get balance: %w", err)
	}

	if balance.Current-sum < 0 {
		return ErrNotEnoughBalance
	}

	_, err = tx.Exec(ctx, `UPDATE balances SET balance = $1, withdrawn = $2 WHERE user_id = $3`,
		balance.Current-sum, balance.Withdrawn+sum, userID)
	if err != nil {
		return fmt.Errorf("can not update balance: %w", err)
	}

	_, err = tx.Exec(ctx, `INSERT INTO withdrawals (user_id, order_id, sum, processed_at)
							VALUES ($1, $2, $3, $4)`, userID, orderNumber, sum, time.Now())
	if err != nil {
		return fmt.Errorf("can not add withdraw: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("can not commit transaction: %w", err)
	}

	return nil
}

func (d *DBRepository) GetWithdrawals(ctx context.Context, userID uuid.UUID) ([]Withdraw, error) {
	rows, err := d.pool.Query(ctx, "SELECT order_id, sum, processed_at FROM withdrawals WHERE user_id = $1", userID)
	if err != nil {
		return nil, fmt.Errorf("can not get orders: %w", err)
	}
	defer rows.Close()

	var withdrawals []Withdraw
	for rows.Next() {
		var withdraw Withdraw
		err = rows.Scan(
			&withdraw.Order,
			&withdraw.Sum,
			&withdraw.ProcessedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("can not read row: %w", err)
		}

		withdrawals = append(withdrawals, withdraw)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("can not read rows: %w", rows.Err())
	}
	if len(withdrawals) == 0 {
		return nil, ErrNoWithdrawals
	}

	return withdrawals, nil
}

func (d *DBRepository) GetOrderForUpdate(ctx context.Context) (*string, *uuid.UUID, error) {
	var order string
	var userID uuid.UUID
	err := d.pool.QueryRow(ctx, `SELECT order_id, user_id
								FROM orders
								WHERE status = $1
								ORDER BY uploaded_at DESC
								LIMIT 1`, models.StatusNew).Scan(&order, &userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, nil
		}
		return nil, nil, fmt.Errorf("can not get order for update: %w", err)
	}

	return &order, &userID, nil
}

func (d *DBRepository) UpdateOrder(
	ctx context.Context,
	orderNumber string,
	status string,
	accrual *float64,
	userID *uuid.UUID,
) error {
	tx, err := d.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("can not start transaction: %w", err)
	}
	defer func() {
		err = tx.Rollback(ctx)
		if err != nil && !strings.Contains(err.Error(), "tx is closed") {
			d.logger.Error("Can not rollback transaction", zap.Error(err))
		}
	}()

	_, err = tx.Exec(ctx, `UPDATE orders SET status = $1, accrual = $2 WHERE order_id = $3`,
		status, accrual, orderNumber)
	if err != nil {
		return fmt.Errorf("can not update order: %w", err)
	}

	if status == string(external.StatusProcessed) && accrual != nil {
		_, err = tx.Exec(ctx, `UPDATE balances SET balance = balance + $1 WHERE user_id = $2`,
			accrual, userID)
		if err != nil {
			return fmt.Errorf("can not update balance: %w", err)
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("can not commit transaction: %w", err)
	}

	return nil
}

func (d *DBRepository) Close() {
	d.pool.Close()
}
