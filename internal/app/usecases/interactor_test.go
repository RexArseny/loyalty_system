package usecases

import (
	"context"
	"encoding/hex"
	"strconv"
	"testing"
	"time"

	"github.com/RexArseny/loyalty_system/internal/app/external"
	"github.com/RexArseny/loyalty_system/internal/app/logger"
	"github.com/RexArseny/loyalty_system/internal/app/models"
	"github.com/RexArseny/loyalty_system/internal/app/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

const (
	testUUIDString      = "3513f2d3-ded4-4305-be83-c0d7ade2508e"
	testLogin           = "testlogin"
	testPassword        = "testpassword"
	testInvalidLogin    = "testinvalidlogin"
	testInvalidPassword = "testinvalidpassword"
	testHash            = "5b7c56a84d56513bd3a469360dc91039539cf9ec41131e19c917943a30164701bca33e0aa7256c13cfdc334579853595b28e02bbfd7b50f069df2fb12439adf9"
	testOrderNumber     = "12345678903"
	testTime            = "2009-11-10T23:00:00Z"
)

var testSalt = []byte{43, 231, 169, 87, 185, 49, 182, 175, 187, 90, 239, 236, 134, 139, 165, 33}

type testRepository struct {
}

func newTestRepository() *testRepository {
	return &testRepository{}
}

func (d *testRepository) Registration(
	_ context.Context,
	_ string,
	_ string,
	_ string,
	_ uuid.UUID,
) error {
	return nil
}

func (d *testRepository) GetUser(_ context.Context, login string) (*repository.User, error) {
	if login == testLogin {
		testUUID, err := uuid.Parse(testUUIDString)
		if err != nil {
			return nil, err
		}

		return &repository.User{
			Login:  testLogin,
			Hash:   testHash,
			Salt:   hex.EncodeToString(testSalt),
			UserID: testUUID,
		}, nil
	}
	return nil, repository.NewErrInvalidAuthData(login)
}

func (d *testRepository) AddOrder(_ context.Context, _ string, _ uuid.UUID) error {
	return nil
}

func (d *testRepository) GetOrders(_ context.Context, _ uuid.UUID) ([]repository.Order, error) {
	testTimeValue, err := time.Parse(time.RFC3339, testTime)
	if err != nil {
		return nil, err
	}
	accrual := float64(100)
	return []repository.Order{
		{
			UploadedAt: testTimeValue,
			Accrual:    &accrual,
			Number:     testOrderNumber,
			Status:     string(models.StatusNew),
		},
	}, nil
}

func (d *testRepository) GetBalance(_ context.Context, _ uuid.UUID) (*repository.Balance, error) {
	return &repository.Balance{
		Current:   100,
		Withdrawn: 200,
	}, nil
}

func (d *testRepository) Withdraw(_ context.Context, _ string, _ float64, _ uuid.UUID) error {
	return nil
}

func (d *testRepository) GetWithdrawals(_ context.Context, _ uuid.UUID) ([]repository.Withdraw, error) {
	testTimeValue, err := time.Parse(time.RFC3339, testTime)
	if err != nil {
		return nil, err
	}
	return []repository.Withdraw{
		{
			ProcessedAt: testTimeValue,
			Order:       testOrderNumber,
			Sum:         100,
		},
	}, nil
}

func (d *testRepository) GetOrdersForUpdate(_ context.Context) ([]repository.Order, error) {
	return []repository.Order{}, nil
}

func (d *testRepository) UpdateOrder(
	_ context.Context,
	_ string,
	_ string,
	_ *float64,
	_ uuid.UUID,
) error {
	return nil
}

func (d *testRepository) Close() {
}

func TestRegistration(t *testing.T) {
	tests := []struct {
		name    string
		request models.AuthRequest
		wantErr bool
	}{
		{
			name: "valid data",
			request: models.AuthRequest{
				Login:    testLogin,
				Password: testPassword,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			testLogger, err := logger.InitLogger()
			assert.NoError(t, err)
			dataRepository := newTestRepository()
			interactor := &Interactor{
				dataRepository:       dataRepository,
				logger:               testLogger.Named("interactor"),
				accrualServiceClient: external.NewAccrualServiceClient(testLogger.Named("accrual"), ""),
			}

			result, err := interactor.Registration(ctx, tt.request)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.NotEmpty(t, result)
		})
	}
}

func TestLogin(t *testing.T) {
	testUUID, err := uuid.Parse(testUUIDString)
	assert.NoError(t, err)
	tests := []struct {
		name    string
		request models.AuthRequest
		want    *uuid.UUID
		wantErr bool
	}{
		{
			name: "invalid data",
			request: models.AuthRequest{
				Login:    testInvalidLogin,
				Password: testInvalidPassword,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "valid data",
			request: models.AuthRequest{
				Login:    testLogin,
				Password: testPassword,
			},
			want:    &testUUID,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			testLogger, err := logger.InitLogger()
			assert.NoError(t, err)
			dataRepository := newTestRepository()
			interactor := &Interactor{
				dataRepository:       dataRepository,
				logger:               testLogger.Named("interactor"),
				accrualServiceClient: external.NewAccrualServiceClient(testLogger.Named("accrual"), ""),
			}

			result, err := interactor.Login(ctx, tt.request)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestAddOrder(t *testing.T) {
	testUUID, err := uuid.Parse(testUUIDString)
	assert.NoError(t, err)
	testOrderNumberValue, err := strconv.Atoi(testOrderNumber)
	assert.NoError(t, err)
	type args struct {
		orderNumber int
		userID      uuid.UUID
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "valid data",
			args: args{
				orderNumber: testOrderNumberValue,
				userID:      testUUID,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		ctx := context.Background()
		testLogger, err := logger.InitLogger()
		assert.NoError(t, err)
		dataRepository := newTestRepository()
		interactor := &Interactor{
			dataRepository:       dataRepository,
			logger:               testLogger.Named("interactor"),
			accrualServiceClient: external.NewAccrualServiceClient(testLogger.Named("accrual"), ""),
		}

		err = interactor.AddOrder(ctx, tt.args.orderNumber, tt.args.userID)
		if tt.wantErr {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
		}
	}
}

func TestGetOrders(t *testing.T) {
	testUUID, err := uuid.Parse(testUUIDString)
	assert.NoError(t, err)
	accrual := float64(100)
	tests := []struct {
		name    string
		userID  uuid.UUID
		want    []models.OrderResponse
		wantErr bool
	}{
		{
			name:   "valid data",
			userID: testUUID,
			want: []models.OrderResponse{
				{
					UploadedAt: testTime,
					Accrual:    &accrual,
					Number:     testOrderNumber,
					Status:     string(models.StatusNew),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			testLogger, err := logger.InitLogger()
			assert.NoError(t, err)
			dataRepository := newTestRepository()
			interactor := &Interactor{
				dataRepository:       dataRepository,
				logger:               testLogger.Named("interactor"),
				accrualServiceClient: external.NewAccrualServiceClient(testLogger.Named("accrual"), ""),
			}

			result, err := interactor.GetOrders(ctx, tt.userID)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestGetBalance(t *testing.T) {
	testUUID, err := uuid.Parse(testUUIDString)
	assert.NoError(t, err)
	tests := []struct {
		name    string
		userID  uuid.UUID
		want    *models.BalanceResponse
		wantErr bool
	}{
		{
			name:   "valid data",
			userID: testUUID,
			want: &models.BalanceResponse{
				Current:   100,
				Withdrawn: 200,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			testLogger, err := logger.InitLogger()
			assert.NoError(t, err)
			dataRepository := newTestRepository()
			interactor := &Interactor{
				dataRepository:       dataRepository,
				logger:               testLogger.Named("interactor"),
				accrualServiceClient: external.NewAccrualServiceClient(testLogger.Named("accrual"), ""),
			}

			result, err := interactor.GetBalance(ctx, tt.userID)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestWithdraw(t *testing.T) {
	testUUID, err := uuid.Parse(testUUIDString)
	assert.NoError(t, err)
	type args struct {
		request models.WithdrawRequest
		userID  uuid.UUID
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "valid data",
			args: args{
				request: models.WithdrawRequest{
					Order: testOrderNumber,
					Sum:   100,
				},
				userID: testUUID,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			testLogger, err := logger.InitLogger()
			assert.NoError(t, err)
			dataRepository := newTestRepository()
			interactor := &Interactor{
				dataRepository:       dataRepository,
				logger:               testLogger.Named("interactor"),
				accrualServiceClient: external.NewAccrualServiceClient(testLogger.Named("accrual"), ""),
			}

			err = interactor.Withdraw(ctx, tt.args.request, tt.args.userID)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetWithdrawals(t *testing.T) {
	testUUID, err := uuid.Parse(testUUIDString)
	assert.NoError(t, err)
	tests := []struct {
		name    string
		userID  uuid.UUID
		want    []models.WithdrawResponse
		wantErr bool
	}{
		{
			name:   "valid data",
			userID: testUUID,
			want: []models.WithdrawResponse{
				{
					Order:       testOrderNumber,
					ProcessedAt: testTime,
					Sum:         100,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			testLogger, err := logger.InitLogger()
			assert.NoError(t, err)
			dataRepository := newTestRepository()
			interactor := &Interactor{
				dataRepository:       dataRepository,
				logger:               testLogger.Named("interactor"),
				accrualServiceClient: external.NewAccrualServiceClient(testLogger.Named("accrual"), ""),
			}

			result, err := interactor.GetWithdrawals(ctx, tt.userID)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestGenerateSalt(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "valid data",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testLogger, err := logger.InitLogger()
			assert.NoError(t, err)
			dataRepository := newTestRepository()
			interactor := &Interactor{
				dataRepository:       dataRepository,
				logger:               testLogger.Named("interactor"),
				accrualServiceClient: external.NewAccrualServiceClient(testLogger.Named("accrual"), ""),
			}

			result, err := interactor.generateSalt()
			assert.NoError(t, err)
			assert.NotEmpty(t, result)
		})
	}
}

func TestHash(t *testing.T) {
	type args struct {
		password []byte
		salt     []byte
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "valid data",
			args: args{
				password: []byte(testPassword),
				salt:     testSalt,
			},
			want: testHash,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testLogger, err := logger.InitLogger()
			assert.NoError(t, err)
			dataRepository := newTestRepository()
			interactor := &Interactor{
				dataRepository:       dataRepository,
				logger:               testLogger.Named("interactor"),
				accrualServiceClient: external.NewAccrualServiceClient(testLogger.Named("accrual"), ""),
			}

			result := interactor.hash(tt.args.password, tt.args.salt)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestChecksum(t *testing.T) {
	testOrderNumberValue, err := strconv.Atoi(testOrderNumber)
	assert.NoError(t, err)
	tests := []struct {
		name   string
		number int
		want   int
	}{
		{
			name:   "valid data",
			number: testOrderNumberValue,
			want:   9,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testLogger, err := logger.InitLogger()
			assert.NoError(t, err)
			dataRepository := newTestRepository()
			interactor := &Interactor{
				dataRepository:       dataRepository,
				logger:               testLogger.Named("interactor"),
				accrualServiceClient: external.NewAccrualServiceClient(testLogger.Named("accrual"), ""),
			}

			result := interactor.checksum(tt.number)
			assert.Equal(t, tt.want, result)
		})
	}
}
