package controllers

import (
	"context"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/RexArseny/loyalty_system/internal/app/external"
	"github.com/RexArseny/loyalty_system/internal/app/logger"
	"github.com/RexArseny/loyalty_system/internal/app/middlewares"
	"github.com/RexArseny/loyalty_system/internal/app/models"
	"github.com/RexArseny/loyalty_system/internal/app/repository"
	"github.com/RexArseny/loyalty_system/internal/app/usecases"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

const (
	testUUIDString  = "3513f2d3-ded4-4305-be83-c0d7ade2508e"
	testLogin       = "testlogin"
	testHash        = "5b7c56a84d56513bd3a469360dc91039539cf9ec41131e19c917943a30164701bca33e0aa7256c13cfdc334579853595b28e02bbfd7b50f069df2fb12439adf9"
	testOrderNumber = "12345678903"
	testTime        = "2009-11-10T23:00:00Z"
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
		name        string
		request     string
		stastusCode int
	}{
		{
			name:        "valid data",
			request:     `{"login":"testlogin", "password":"testpassword"}`,
			stastusCode: http.StatusOK,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testLogger, err := logger.InitLogger()
			assert.NoError(t, err)
			dataRepository := newTestRepository()
			interactor := usecases.NewInteractor(
				context.Background(),
				testLogger.Named("interactor"),
				dataRepository,
				external.NewAccrualServiceClient(testLogger.Named("accrual"), ""),
			)
			conntroller := NewController(testLogger.Named("controller"), interactor)
			middleware, err := middlewares.NewMiddleware(
				"../../../public.pem",
				"../../../private.pem",
				testLogger.Named("middleware"),
			)
			assert.NoError(t, err)

			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = httptest.NewRequest(http.MethodPost, "/api/user/register", strings.NewReader(tt.request))

			conntroller.Registration(ctx)

			auth := middleware.SetJWT()
			auth(ctx)

			result := w.Result()

			assert.Equal(t, tt.stastusCode, result.StatusCode)

			w.Result().Body.Close()
		})
	}
}

func TestLogin(t *testing.T) {
	tests := []struct {
		name        string
		request     string
		stastusCode int
	}{
		{
			name:        "valid data",
			request:     `{"login":"testlogin", "password":"testpassword"}`,
			stastusCode: http.StatusOK,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testLogger, err := logger.InitLogger()
			assert.NoError(t, err)
			dataRepository := newTestRepository()
			interactor := usecases.NewInteractor(
				context.Background(),
				testLogger.Named("interactor"),
				dataRepository,
				external.NewAccrualServiceClient(testLogger.Named("accrual"), ""),
			)
			conntroller := NewController(testLogger.Named("controller"), interactor)
			middleware, err := middlewares.NewMiddleware(
				"../../../public.pem",
				"../../../private.pem",
				testLogger.Named("middleware"),
			)
			assert.NoError(t, err)

			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = httptest.NewRequest(http.MethodPost, "/api/user/login", strings.NewReader(tt.request))

			conntroller.Login(ctx)

			auth := middleware.SetJWT()
			auth(ctx)

			result := w.Result()

			assert.Equal(t, tt.stastusCode, result.StatusCode)

			w.Result().Body.Close()
		})
	}
}

func TestAddOrder(t *testing.T) {
	tests := []struct {
		name         string
		loginRequest string
		request      string
		stastusCode  int
	}{
		{
			name:         "valid data",
			loginRequest: `{"login":"testlogin", "password":"testpassword"}`,
			request:      testOrderNumber,
			stastusCode:  http.StatusAccepted,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testLogger, err := logger.InitLogger()
			assert.NoError(t, err)
			dataRepository := newTestRepository()
			interactor := usecases.NewInteractor(
				context.Background(),
				testLogger.Named("interactor"),
				dataRepository,
				external.NewAccrualServiceClient(testLogger.Named("accrual"), ""),
			)
			conntroller := NewController(testLogger.Named("controller"), interactor)
			middleware, err := middlewares.NewMiddleware(
				"../../../public.pem",
				"../../../private.pem",
				testLogger.Named("middleware"),
			)
			assert.NoError(t, err)

			wLogin := httptest.NewRecorder()
			ctxLogin, _ := gin.CreateTestContext(wLogin)
			ctxLogin.Request = httptest.NewRequest(http.MethodPost, "/api/user/login", strings.NewReader(tt.loginRequest))

			conntroller.Login(ctxLogin)

			authLogin := middleware.SetJWT()
			authLogin(ctxLogin)

			resultLogin := wLogin.Result()

			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = httptest.NewRequest(http.MethodPost, "/api/user/orders", strings.NewReader(tt.request))
			for _, cookie := range resultLogin.Cookies() {
				if cookie != nil && cookie.Name == middlewares.Authorization {
					ctx.Request.AddCookie(&http.Cookie{
						Name:  middlewares.Authorization,
						Value: cookie.Value,
					})
					break
				}
			}
			auth := middleware.GetJWT()
			auth(ctx)

			conntroller.AddOrder(ctx)

			result := w.Result()

			assert.Equal(t, tt.stastusCode, result.StatusCode)

			wLogin.Result().Body.Close()
			w.Result().Body.Close()
		})
	}
}

func TestGetOrders(t *testing.T) {
	tests := []struct {
		name         string
		loginRequest string
		stastusCode  int
	}{
		{
			name:         "valid data",
			loginRequest: `{"login":"testlogin", "password":"testpassword"}`,
			stastusCode:  http.StatusOK,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testLogger, err := logger.InitLogger()
			assert.NoError(t, err)
			dataRepository := newTestRepository()
			interactor := usecases.NewInteractor(
				context.Background(),
				testLogger.Named("interactor"),
				dataRepository,
				external.NewAccrualServiceClient(testLogger.Named("accrual"), ""),
			)
			conntroller := NewController(testLogger.Named("controller"), interactor)
			middleware, err := middlewares.NewMiddleware(
				"../../../public.pem",
				"../../../private.pem",
				testLogger.Named("middleware"),
			)
			assert.NoError(t, err)

			wLogin := httptest.NewRecorder()
			ctxLogin, _ := gin.CreateTestContext(wLogin)
			ctxLogin.Request = httptest.NewRequest(http.MethodPost, "/api/user/login", strings.NewReader(tt.loginRequest))

			conntroller.Login(ctxLogin)

			authLogin := middleware.SetJWT()
			authLogin(ctxLogin)

			resultLogin := wLogin.Result()

			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)
			for _, cookie := range resultLogin.Cookies() {
				if cookie != nil && cookie.Name == middlewares.Authorization {
					ctx.Request.AddCookie(&http.Cookie{
						Name:  middlewares.Authorization,
						Value: cookie.Value,
					})
					break
				}
			}
			auth := middleware.GetJWT()
			auth(ctx)

			conntroller.GetOrders(ctx)

			result := w.Result()

			assert.Equal(t, tt.stastusCode, result.StatusCode)

			wLogin.Result().Body.Close()
			w.Result().Body.Close()
		})
	}
}

func TestGetBalance(t *testing.T) {
	tests := []struct {
		name         string
		loginRequest string
		stastusCode  int
	}{
		{
			name:         "valid data",
			loginRequest: `{"login":"testlogin", "password":"testpassword"}`,
			stastusCode:  http.StatusOK,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testLogger, err := logger.InitLogger()
			assert.NoError(t, err)
			dataRepository := newTestRepository()
			interactor := usecases.NewInteractor(
				context.Background(),
				testLogger.Named("interactor"),
				dataRepository,
				external.NewAccrualServiceClient(testLogger.Named("accrual"), ""),
			)
			conntroller := NewController(testLogger.Named("controller"), interactor)
			middleware, err := middlewares.NewMiddleware(
				"../../../public.pem",
				"../../../private.pem",
				testLogger.Named("middleware"),
			)
			assert.NoError(t, err)

			wLogin := httptest.NewRecorder()
			ctxLogin, _ := gin.CreateTestContext(wLogin)
			ctxLogin.Request = httptest.NewRequest(http.MethodPost, "/api/user/login", strings.NewReader(tt.loginRequest))

			conntroller.Login(ctxLogin)

			authLogin := middleware.SetJWT()
			authLogin(ctxLogin)

			resultLogin := wLogin.Result()

			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = httptest.NewRequest(http.MethodGet, "/api/user/balance", nil)
			for _, cookie := range resultLogin.Cookies() {
				if cookie != nil && cookie.Name == middlewares.Authorization {
					ctx.Request.AddCookie(&http.Cookie{
						Name:  middlewares.Authorization,
						Value: cookie.Value,
					})
					break
				}
			}
			auth := middleware.GetJWT()
			auth(ctx)

			conntroller.GetBalance(ctx)

			result := w.Result()

			assert.Equal(t, tt.stastusCode, result.StatusCode)

			wLogin.Result().Body.Close()
			w.Result().Body.Close()
		})
	}
}

func TestWithdraw(t *testing.T) {
	tests := []struct {
		name         string
		loginRequest string
		request      string
		stastusCode  int
	}{
		{
			name:         "valid data",
			loginRequest: `{"login":"testlogin", "password":"testpassword"}`,
			request:      `{"order":"12345678903", "sum":100}`,
			stastusCode:  http.StatusOK,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testLogger, err := logger.InitLogger()
			assert.NoError(t, err)
			dataRepository := newTestRepository()
			interactor := usecases.NewInteractor(
				context.Background(),
				testLogger.Named("interactor"),
				dataRepository,
				external.NewAccrualServiceClient(testLogger.Named("accrual"), ""),
			)
			conntroller := NewController(testLogger.Named("controller"), interactor)
			middleware, err := middlewares.NewMiddleware(
				"../../../public.pem",
				"../../../private.pem",
				testLogger.Named("middleware"),
			)
			assert.NoError(t, err)

			wLogin := httptest.NewRecorder()
			ctxLogin, _ := gin.CreateTestContext(wLogin)
			ctxLogin.Request = httptest.NewRequest(http.MethodPost, "/api/user/login", strings.NewReader(tt.loginRequest))

			conntroller.Login(ctxLogin)

			authLogin := middleware.SetJWT()
			authLogin(ctxLogin)

			resultLogin := wLogin.Result()

			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", strings.NewReader(tt.request))
			for _, cookie := range resultLogin.Cookies() {
				if cookie != nil && cookie.Name == middlewares.Authorization {
					ctx.Request.AddCookie(&http.Cookie{
						Name:  middlewares.Authorization,
						Value: cookie.Value,
					})
					break
				}
			}
			auth := middleware.GetJWT()
			auth(ctx)

			conntroller.Withdraw(ctx)

			result := w.Result()

			assert.Equal(t, tt.stastusCode, result.StatusCode)

			wLogin.Result().Body.Close()
			w.Result().Body.Close()
		})
	}
}

func TestGetWithdrawals(t *testing.T) {
	tests := []struct {
		name         string
		loginRequest string
		stastusCode  int
	}{
		{
			name:         "valid data",
			loginRequest: `{"login":"testlogin", "password":"testpassword"}`,
			stastusCode:  http.StatusOK,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testLogger, err := logger.InitLogger()
			assert.NoError(t, err)
			dataRepository := newTestRepository()
			interactor := usecases.NewInteractor(
				context.Background(),
				testLogger.Named("interactor"),
				dataRepository,
				external.NewAccrualServiceClient(testLogger.Named("accrual"), ""),
			)
			conntroller := NewController(testLogger.Named("controller"), interactor)
			middleware, err := middlewares.NewMiddleware(
				"../../../public.pem",
				"../../../private.pem",
				testLogger.Named("middleware"),
			)
			assert.NoError(t, err)

			wLogin := httptest.NewRecorder()
			ctxLogin, _ := gin.CreateTestContext(wLogin)
			ctxLogin.Request = httptest.NewRequest(http.MethodPost, "/api/user/login", strings.NewReader(tt.loginRequest))

			conntroller.Login(ctxLogin)

			authLogin := middleware.SetJWT()
			authLogin(ctxLogin)

			resultLogin := wLogin.Result()

			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = httptest.NewRequest(http.MethodGet, "/api/user/withdrawals", nil)
			for _, cookie := range resultLogin.Cookies() {
				if cookie != nil && cookie.Name == middlewares.Authorization {
					ctx.Request.AddCookie(&http.Cookie{
						Name:  middlewares.Authorization,
						Value: cookie.Value,
					})
					break
				}
			}
			auth := middleware.GetJWT()
			auth(ctx)

			conntroller.GetWithdrawals(ctx)

			result := w.Result()

			assert.Equal(t, tt.stastusCode, result.StatusCode)

			wLogin.Result().Body.Close()
			w.Result().Body.Close()
		})
	}
}
