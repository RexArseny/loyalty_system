package routers

import (
	"github.com/RexArseny/loyalty_system/internal/app/config"
	"github.com/RexArseny/loyalty_system/internal/app/controllers"
	"github.com/RexArseny/loyalty_system/internal/app/middlewares"
	"github.com/gin-gonic/gin"
)

func NewRouter(
	cfg *config.Config,
	controller controllers.Controller,
	middleware *middlewares.Middleware,
) (*gin.Engine, error) {
	router := gin.New()
	router.Use(
		gin.Recovery(),
		middleware.Logger(),
	)

	groupWithoutJWT := router.Group("", middleware.SetJWT())
	{
		groupWithoutJWT.POST("/api/user/register", controller.Registration)
		groupWithoutJWT.POST("/api/user/login", controller.Login)
	}

	groupWithJWT := router.Group("", middleware.GetJWT())
	{
		groupWithJWT.POST("/api/user/orders", controller.AddOrder)
		groupWithJWT.GET("/api/user/orders", controller.GetOrders)
		groupWithJWT.GET("/api/user/balance", controller.GetBalance)
		groupWithJWT.POST("/api/user/balance/withdraw", controller.Withdraw)
		groupWithJWT.GET("/api/user/withdrawals", controller.GetWithdrawals)
	}

	return router, nil
}
