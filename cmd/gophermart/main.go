// Package main является точкой входа приложения Gophermart —
// системы лояльности с накоплением баллов за заказы.
package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/eshadow1/gophermart/internal/client"
	"github.com/eshadow1/gophermart/internal/configs"
	"github.com/eshadow1/gophermart/internal/handler"
	"github.com/eshadow1/gophermart/internal/loggers"
	"github.com/eshadow1/gophermart/internal/repository"
	"github.com/eshadow1/gophermart/internal/service"
	"github.com/eshadow1/gophermart/internal/worker"
)

// Таймауты HTTP-сервера для защиты от медленных клиентов и корректной работы.
const (
	// defaultReadTimeout — максимальное время чтения всего HTTP-запроса,
	// включая тело.
	defaultReadTimeout = 15 * time.Second
	// defaultWriteTimeout — максимальное время записи HTTP-ответа.
	defaultWriteTimeout = 15 * time.Second
	// defaultIdleTimeout — максимальное время простоя keep-alive соединения.
	defaultIdleTimeout = 60 * time.Second
	// defaultShutdownTimeout — максимальное время, отводимое на graceful shutdown
	// сервера и фоновых процессов.
	defaultShutdownTimeout = 30 * time.Second
)

// main — точка входа приложения. Выполняет инициализацию всех компонентов,
// запуск HTTP-сервера и фонового воркера, ожидание сигнала завершения
// и graceful shutdown.
func main() {
	cfg := configs.NewConfig()
	cfg.Init()

	errCreateLog := loggers.CreateLogger(cfg.Log.Level)
	if errCreateLog != nil {
		fmt.Println("Error creating logger:", errCreateLog)
		return
	}

	ctxWorker, cancelWorker := context.WithCancel(context.Background())
	defer cancelWorker()

	builder := repository.NewBuilderConnect()

	dbConnect, errConnect := builder.CreateNewPool(&cfg.Storage)
	if errConnect != nil {
		loggers.Log.Fatalf("error connecting to database: %v", errConnect)
	}

	orderR := repository.NewOrderRepo(dbConnect)
	orderS := service.NewOrderService(cfg, orderR)
	orderH := handler.NewOrderHandler(cfg, orderS)

	balanceR := repository.NewBalanceRepo(dbConnect)
	balanceS := service.NewBalanceService(cfg, balanceR)
	balanceH := handler.NewBalanceHandler(cfg, balanceS)

	userR := repository.NewUserRepo(dbConnect)
	authS := service.NewAuthService(&cfg.Auth, userR)
	authH := handler.NewAuthHandler(authS)

	rs := handler.InitRouter(cfg, orderH, balanceH, authH)

	server := &http.Server{
		Addr:         cfg.Addr,
		Handler:      rs,
		ReadTimeout:  defaultReadTimeout,
		WriteTimeout: defaultWriteTimeout,
		IdleTimeout:  defaultIdleTimeout,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			loggers.Log.Fatalf("Server failed: %v", err)
		}
	}()

	maxConcurrent := 10
	batchSize := 50
	interval := 5 * time.Second
	accClient := client.NewAccrualClient(cfg.AddrSystem)
	accWorker := worker.NewAccrualWorker(orderR, balanceR, accClient, maxConcurrent, batchSize, interval)
	go accWorker.Run(ctxWorker)

	<-quit
	loggers.Log.Infoln("Shutting down server...")

	cancelWorker()

	ctx, cancel := context.WithTimeout(context.Background(), defaultShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		loggers.Log.Infof("Server forced to shutdown: %v", err)
		return
	}
}
