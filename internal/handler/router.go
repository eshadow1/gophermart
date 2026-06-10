package handler

import (
	"net/http"
	"time"

	"github.com/eshadow1/gophermart/internal/configs"
	customMiddleware "github.com/eshadow1/gophermart/internal/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const (
	timeoutRequest = 5 * time.Second
)

// AuthHandler определяет контракт для HTTP-хендлеров, отвечающих
// за регистрацию и аутентификацию пользователей.
type AuthHandler interface {
	RegisterUser(w http.ResponseWriter, r *http.Request)
	LoginUser(w http.ResponseWriter, r *http.Request)
}

// Order определяет контракт для HTTP-хендлеров, отвечающих
// за загрузку и получение списка заказов пользователя.
type Order interface {
	UploadUserOrder(w http.ResponseWriter, r *http.Request)
	GetUserOrders(w http.ResponseWriter, r *http.Request)
}

// Balance определяет контракт для HTTP-хендлеров, отвечающих
// за проверку баланса, вывод средств и получение истории выводов.
type Balance interface {
	GetInfoWithdrawals(w http.ResponseWriter, r *http.Request)
	GetBalanceUser(w http.ResponseWriter, r *http.Request)
	RequestWithdraw(w http.ResponseWriter, r *http.Request)
}

// InitRouter инициализирует и настраивает HTTP-маршрутизатор (chi.Mux) для приложения.
func InitRouter(cfg *configs.Config, o Order, b Balance, auth AuthHandler) *chi.Mux {
	rs := chi.NewRouter()
	rs.Use(customMiddleware.GzipMiddleware(), customMiddleware.LoggerMiddleware(), middleware.Timeout(timeoutRequest))
	rs.Route("/api/user/", func(r chi.Router) {
		r.Post("/register", auth.RegisterUser)
		r.Post("/login", auth.LoginUser)
		r.Route("/", func(r chi.Router) {
			r.Use(customMiddleware.AuthMiddleware(&cfg.Auth))
			r.Post("/orders", o.UploadUserOrder)
			r.Get("/orders", o.GetUserOrders)
			r.Get("/withdrawals", b.GetInfoWithdrawals)
			r.Route("/balance", func(r chi.Router) {
				r.Get("/", b.GetBalanceUser)
				r.Post("/withdraw", b.RequestWithdraw)
			})
		})
	})

	return rs
}
