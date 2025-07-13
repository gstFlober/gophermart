package main

import (
	"context"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/rs/zerolog"
	"gophemart/internal/app/service"
	"gophemart/internal/config"
	"gophemart/internal/handler/http"
	"gophemart/internal/repository/postgresql"
	"gophemart/internal/transport/accrual"
	"gophemart/internal/worker"
	"gophemart/pkg/database"
	"gophemart/pkg/jwt"
	"gophemart/pkg/logger"
	"log"
	n "net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {

	logger.Init(zerolog.DebugLevel)

	cfg := config.MustLoad()
	db, err := database.NewPostgresDB(cfg.Database)
	if err != nil {
		return
	}
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	if err := database.Migrate(db); err != nil {
		logger.Error().
			Err(err).
			Str("handler", "GetWithdrawals").
			Msg("Failed to get user withdrawals")
		return
	}
	repo := postgresql.NewRepository(db)
	jwtManager := jwt.NewManager(cfg.Auth.JWTSecret, 30*24*time.Hour)
	accrualClient := accrual.NewClient(cfg.Accural)

	log.Println("cfg.Accural = ", cfg.Accural)
	authService := service.NewAuthService(repo.User, cfg.Auth.JWTSecret)
	orderService := service.NewOrderService(repo.Order, repo.User, accrualClient)
	balanceService := service.NewBalanceService(repo.User, repo.Order, repo.Withdrawal)

	authHandler := http.NewAuthHandler(authService, jwtManager)
	orderHandler := http.NewOrderHandler(orderService)
	balanceHandler := http.NewBalanceHandler(balanceService)

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"http://localhost", "*"},
		AllowMethods:     []string{n.MethodGet, n.MethodPost, n.MethodPut, n.MethodDelete},
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization, echo.HeaderCookie},
		AllowCredentials: true,

		MaxAge: 86400,
	}))
	api := e.Group("/api")

	api.POST("/user/register", authHandler.Register)
	api.POST("/user/login", authHandler.Login)
	authGroup := api.Group("")

	authGroup.Use(http.AuthMiddleware(jwtManager))

	authGroup.POST("/user/orders", orderHandler.UploadOrder)
	authGroup.GET("/user/orders", orderHandler.GetOrders)
	authGroup.GET("/user/balance", balanceHandler.GetBalance)
	authGroup.POST("/user/balance/withdraw", balanceHandler.Withdraw)
	authGroup.GET("/user/withdrawals", balanceHandler.GetWithdrawals)
	for _, route := range e.Routes() {
		log.Printf("Registered: %-6s %s", route.Method, route.Path)
	}

	orderProcessor := worker.NewOrderProcessor(
		repo.Order,
		repo.User,
		accrualClient,
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go orderProcessor.Run(ctx, 5*time.Second)

	go func() {
		if err := e.Start(cfg.Server.Address); err != nil {
			log.Println("QQQ ", err)
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	<-quit

	ctx2, cancel2 := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)

	defer cancel2()
	if err := e.Shutdown(ctx2); err != nil {
		log.Println("QQQ ", err)
	}

}
