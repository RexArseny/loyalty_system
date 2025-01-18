package main

import (
	"context"
	"log"

	"github.com/RexArseny/loyalty_system/internal/app"
	"github.com/RexArseny/loyalty_system/internal/app/config"
	"github.com/RexArseny/loyalty_system/internal/app/logger"
	"github.com/RexArseny/loyalty_system/internal/app/repository"
	"go.uber.org/zap"
)

func main() {
	ctx := context.Background()

	mainLogger, err := logger.InitLogger()
	if err != nil {
		log.Fatalf("Can not init logger: %s", err)
	}
	defer func() {
		if err := mainLogger.Sync(); err != nil {
			log.Fatalf("Logger sync failed: %s", err)
		}
	}()

	cfg, err := config.Init()
	if err != nil {
		mainLogger.Fatal("Can not init config", zap.Error(err))
	}

	dataRepository, err := repository.NewRepository(
		ctx,
		mainLogger.Named("repository"),
		cfg.DatabaseURI,
	)
	if err != nil {
		mainLogger.Fatal("Can not init repository", zap.Error(err))
	}
	defer func() {
		dataRepository.Close()
	}()

	s, err := app.NewServer(ctx, mainLogger, cfg, dataRepository)
	if err != nil {
		mainLogger.Fatal("Can not init server", zap.Error(err))
	}

	err = s.ListenAndServe()
	if err != nil {
		mainLogger.Fatal("Can not listen and serve", zap.Error(err))
	}
}
