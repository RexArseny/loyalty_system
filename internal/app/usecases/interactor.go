package usecases

import (
	"context"
	"crypto/rand"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"github.com/RexArseny/loyalty_system/internal/app/external"
	"github.com/RexArseny/loyalty_system/internal/app/models"
	"github.com/RexArseny/loyalty_system/internal/app/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	saltSize         = 16
	statusCheckTimer = 100
)

type Interactor struct {
	dataRepository       repository.Repository
	logger               *zap.Logger
	accrualServiceClient external.AccrualServiceClient
}

func NewInteractor(
	ctx context.Context,
	logger *zap.Logger,
	dataRepository repository.Repository,
	accrualServiceClient external.AccrualServiceClient,
) Interactor {
	interactor := Interactor{
		dataRepository:       dataRepository,
		accrualServiceClient: accrualServiceClient,
		logger:               logger,
	}

	go interactor.runStatusCheck(ctx)

	return interactor
}

func (i *Interactor) runStatusCheck(ctx context.Context) {
	ticker := time.NewTicker(statusCheckTimer * time.Millisecond)
	for range ticker.C {
		order, userID, err := i.dataRepository.GetOrderForUpdate(ctx)
		if err != nil {
			i.logger.Error("Can not get order for update", zap.Error(err))
			return
		}
		if order == nil {
			continue
		}
		data, err := i.accrualServiceClient.GetData(ctx, order)
		if err != nil {
			i.logger.Error("Can not get data from accrual service", zap.Error(err))
			return
		}
		err = i.dataRepository.UpdateOrder(ctx, data.Order, string(data.Status), data.Accrual, userID)
		if err != nil {
			i.logger.Error("Can not update order", zap.Error(err))
			return
		}
	}
}

func (i *Interactor) Registration(ctx context.Context, request models.AuthRequest) (*uuid.UUID, error) {
	userID := uuid.New()

	salt, err := i.generateSalt()
	if err != nil {
		return nil, fmt.Errorf("can not generate salt: %w", err)
	}
	hash := i.hash([]byte(request.Password), salt)

	err = i.dataRepository.Registration(ctx, request.Login, hash, hex.EncodeToString(salt), userID)
	if err != nil {
		return nil, fmt.Errorf("can not register user: %w", err)
	}

	return &userID, nil
}

func (i *Interactor) Login(ctx context.Context, request models.AuthRequest) (*uuid.UUID, error) {
	data, err := i.dataRepository.GetUser(ctx, request.Login)
	if err != nil {
		return nil, fmt.Errorf("can not get login and password: %w", err)
	}

	salt, err := hex.DecodeString(data.Salt)
	if err != nil {
		return nil, fmt.Errorf("can not decode salt: %w", err)
	}
	if i.hash([]byte(request.Password), salt) != data.Hash {
		return nil, repository.ErrInvalidAuthData
	}

	return &data.UserID, nil
}

func (i *Interactor) AddOrder(ctx context.Context, orderNumber int, userID uuid.UUID) error {
	if (orderNumber%10+i.checksum(orderNumber/10))%10 != 0 {
		return repository.ErrInvalidOrderNumber
	}

	err := i.dataRepository.AddOrder(ctx, strconv.Itoa(orderNumber), userID)
	if err != nil {
		return fmt.Errorf("can not add order: %w", err)
	}

	return nil
}

func (i *Interactor) GetOrders(ctx context.Context, userID uuid.UUID) ([]models.OrderResponse, error) {
	data, err := i.dataRepository.GetOrders(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("can not get orders: %w", err)
	}

	response := make([]models.OrderResponse, 0, len(data))
	for _, item := range data {
		response = append(response, models.OrderResponse{
			Number:     item.Number,
			Status:     item.Status,
			Accrual:    item.Accrual,
			UploadedAt: item.UploadedAt.Format(time.RFC3339),
		})
	}

	return response, nil
}

func (i *Interactor) GetBalance(ctx context.Context, userID uuid.UUID) (*models.BalanceResponse, error) {
	data, err := i.dataRepository.GetBalance(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("can not get balance: %w", err)
	}

	return &models.BalanceResponse{
		Current:   data.Current,
		Withdrawn: data.Withdrawn,
	}, nil
}

func (i *Interactor) Withdraw(ctx context.Context, request models.WithdrawRequest, userID uuid.UUID) error {
	orderNumber, err := strconv.Atoi(request.Order)
	if err != nil {
		return fmt.Errorf("can not parse order number: %w", err)
	}
	if (orderNumber%10+i.checksum(orderNumber/10))%10 != 0 {
		return repository.ErrInvalidOrderNumber
	}

	err = i.dataRepository.Withdraw(ctx, request.Order, request.Sum, userID)
	if err != nil {
		return fmt.Errorf("can not withdraw: %w", err)
	}

	return nil
}

func (i *Interactor) GetWithdrawals(ctx context.Context, userID uuid.UUID) ([]models.WithdrawResponse, error) {
	data, err := i.dataRepository.GetWithdrawals(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("can not get withdrawals: %w", err)
	}

	response := make([]models.WithdrawResponse, 0, len(data))
	for _, item := range data {
		response = append(response, models.WithdrawResponse{
			Order:       item.Order,
			Sum:         item.Sum,
			ProcessedAt: item.ProcessedAt.Format(time.RFC3339),
		})
	}

	return response, nil
}

func (i *Interactor) generateSalt() ([]byte, error) {
	salt := make([]byte, saltSize)

	_, err := rand.Read(salt)
	if err != nil {
		return nil, fmt.Errorf("can not generate salt: %w", err)
	}

	return salt, nil
}

func (i *Interactor) hash(password []byte, salt []byte) string {
	hash := sha512.New()
	hash.Write(append(password, salt...))
	hashedPassword := hash.Sum(nil)

	return hex.EncodeToString(hashedPassword)
}

func (i *Interactor) checksum(number int) int {
	var luhn int

	for i := 0; number > 0; i++ {
		cur := number % 10

		if i%2 == 0 {
			cur = cur * 2
			if cur > 9 {
				cur = cur%10 + cur/10
			}
		}

		luhn += cur
		number = number / 10
	}

	return luhn % 10
}
