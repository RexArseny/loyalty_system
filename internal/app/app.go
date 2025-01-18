package app

import (
	"context"
	"fmt"
	"net/http"

	"github.com/RexArseny/loyalty_system/internal/app/config"
	"github.com/RexArseny/loyalty_system/internal/app/controllers"
	"github.com/RexArseny/loyalty_system/internal/app/external"
	"github.com/RexArseny/loyalty_system/internal/app/middlewares"
	"github.com/RexArseny/loyalty_system/internal/app/repository"
	"github.com/RexArseny/loyalty_system/internal/app/routers"
	"github.com/RexArseny/loyalty_system/internal/app/usecases"
	"go.uber.org/zap"
)

func NewServer(
	ctx context.Context,
	logger *zap.Logger,
	cfg *config.Config,
	dataRepository repository.Repository,
) (*http.Server, error) {
	accrualServiceClient := external.NewAccrualServiceClient(cfg.AccrualSystemAddress)
	interactor := usecases.NewInteractor(ctx, logger.Named("interactor"), dataRepository, accrualServiceClient)
	controller := controllers.NewController(logger.Named("controller"), interactor)
	middleware, err := middlewares.NewMiddleware(
		cfg.PublicKeyPath,
		cfg.PrivateKeyPath,
		logger.Named("middleware"),
	)
	if err != nil {
		return nil, fmt.Errorf("can not init middleware: %w", err)
	}
	router, err := routers.NewRouter(cfg, controller, middleware)
	if err != nil {
		return nil, fmt.Errorf("can not init router: %w", err)
	}

	return &http.Server{
		Addr:    cfg.RunAddress,
		Handler: router,
	}, nil
}
