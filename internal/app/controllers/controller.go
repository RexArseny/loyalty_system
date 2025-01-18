//nolint:dupl // auth metods are grouped
package controllers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/RexArseny/loyalty_system/internal/app/middlewares"
	"github.com/RexArseny/loyalty_system/internal/app/models"
	"github.com/RexArseny/loyalty_system/internal/app/repository"
	"github.com/RexArseny/loyalty_system/internal/app/usecases"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Controller struct {
	logger     *zap.Logger
	interactor usecases.Interactor
}

func NewController(logger *zap.Logger, interactor usecases.Interactor) Controller {
	return Controller{
		logger:     logger,
		interactor: interactor,
	}
}

func (c *Controller) Registration(ctx *gin.Context) {
	data, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": http.StatusText(http.StatusBadRequest)})
		return
	}

	var request models.AuthRequest
	err = json.Unmarshal(data, &request)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": http.StatusText(http.StatusBadRequest)})
		return
	}

	result, err := c.interactor.Registration(ctx, request)
	if err != nil {
		if errors.Is(err, repository.ErrOriginalLoginUniqueViolation) {
			ctx.JSON(http.StatusConflict, gin.H{"error": http.StatusText(http.StatusConflict)})
			return
		}
		c.logger.Error("Can not register user", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": http.StatusText(http.StatusInternalServerError)})
		return
	}

	ctx.Set(middlewares.UserID, result)
}

func (c *Controller) Login(ctx *gin.Context) {
	data, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": http.StatusText(http.StatusBadRequest)})
		return
	}

	var request models.AuthRequest
	err = json.Unmarshal(data, &request)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": http.StatusText(http.StatusBadRequest)})
		return
	}

	result, err := c.interactor.Login(ctx, request)
	if err != nil {
		if errors.Is(err, repository.ErrInvalidAuthData) {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": http.StatusText(http.StatusUnauthorized)})
			return
		}
		c.logger.Error("Can not login user", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": http.StatusText(http.StatusInternalServerError)})
		return
	}

	ctx.Set(middlewares.UserID, result)
}

func (c *Controller) AddOrder(ctx *gin.Context) {
	tokenValue, ok := ctx.Get(middlewares.Authorization)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": http.StatusText(http.StatusUnauthorized)})
		return
	}
	token, ok := tokenValue.(*middlewares.JWT)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": http.StatusText(http.StatusUnauthorized)})
		return
	}

	data, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": http.StatusText(http.StatusBadRequest)})
		return
	}

	var request int
	err = json.Unmarshal(data, &request)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": http.StatusText(http.StatusBadRequest)})
		return
	}

	err = c.interactor.AddOrder(ctx, request, token.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrAlreadyAdded) {
			ctx.JSON(http.StatusOK, gin.H{"error": http.StatusText(http.StatusOK)})
			return
		}
		if errors.Is(err, repository.ErrAlreadyAddedByAnotherUser) {
			ctx.JSON(http.StatusConflict, gin.H{"error": http.StatusText(http.StatusConflict)})
			return
		}
		if errors.Is(err, repository.ErrInvalidOrderNumber) {
			ctx.JSON(http.StatusUnprocessableEntity, gin.H{"error": http.StatusText(http.StatusUnprocessableEntity)})
			return
		}
		c.logger.Error("Can not add order", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": http.StatusText(http.StatusInternalServerError)})
		return
	}

	ctx.JSON(http.StatusAccepted, gin.H{"status": http.StatusText(http.StatusAccepted)})
}

func (c *Controller) GetOrders(ctx *gin.Context) {
	tokenValue, ok := ctx.Get(middlewares.Authorization)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": http.StatusText(http.StatusUnauthorized)})
		return
	}
	token, ok := tokenValue.(*middlewares.JWT)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": http.StatusText(http.StatusUnauthorized)})
		return
	}

	result, err := c.interactor.GetOrders(ctx, token.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrNoOrders) {
			ctx.JSON(http.StatusNoContent, gin.H{"error": http.StatusText(http.StatusNoContent)})
			return
		}
		c.logger.Error("Can not get orders", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": http.StatusText(http.StatusInternalServerError)})
		return
	}

	ctx.JSON(http.StatusOK, result)
}

func (c *Controller) GetBalance(ctx *gin.Context) {
	tokenValue, ok := ctx.Get(middlewares.Authorization)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": http.StatusText(http.StatusUnauthorized)})
		return
	}
	token, ok := tokenValue.(*middlewares.JWT)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": http.StatusText(http.StatusUnauthorized)})
		return
	}

	result, err := c.interactor.GetBalance(ctx, token.UserID)
	if err != nil {
		c.logger.Error("Can not get balance", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": http.StatusText(http.StatusInternalServerError)})
		return
	}

	ctx.JSON(http.StatusOK, result)
}

func (c *Controller) Withdraw(ctx *gin.Context) {
	tokenValue, ok := ctx.Get(middlewares.Authorization)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": http.StatusText(http.StatusUnauthorized)})
		return
	}
	token, ok := tokenValue.(*middlewares.JWT)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": http.StatusText(http.StatusUnauthorized)})
		return
	}

	data, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": http.StatusText(http.StatusBadRequest)})
		return
	}

	var request models.WithdrawRequest
	err = json.Unmarshal(data, &request)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": http.StatusText(http.StatusBadRequest)})
		return
	}

	err = c.interactor.Withdraw(ctx, request, token.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrNotEnoughBalance) {
			ctx.JSON(http.StatusPaymentRequired, gin.H{"error": http.StatusText(http.StatusPaymentRequired)})
			return
		}
		if errors.Is(err, repository.ErrInvalidOrderNumber) {
			ctx.JSON(http.StatusUnprocessableEntity, gin.H{"error": http.StatusText(http.StatusUnprocessableEntity)})
			return
		}
		c.logger.Error("Can not withdraw", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": http.StatusText(http.StatusInternalServerError)})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": http.StatusText(http.StatusOK)})
}

func (c *Controller) GetWithdrawals(ctx *gin.Context) {
	tokenValue, ok := ctx.Get(middlewares.Authorization)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": http.StatusText(http.StatusUnauthorized)})
		return
	}
	token, ok := tokenValue.(*middlewares.JWT)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": http.StatusText(http.StatusUnauthorized)})
		return
	}

	result, err := c.interactor.GetWithdrawals(ctx, token.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrNoWithdrawals) {
			ctx.JSON(http.StatusNoContent, gin.H{"error": http.StatusText(http.StatusNoContent)})
			return
		}
		c.logger.Error("Can not get withdrawals", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": http.StatusText(http.StatusInternalServerError)})
		return
	}

	ctx.JSON(http.StatusOK, result)
}
