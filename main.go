package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"bankAPI/internal/handler"
	"bankAPI/internal/middleware"
	"bankAPI/internal/repository"
	"bankAPI/internal/service"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

func main() {
	// Загрузка .env
	if err := godotenv.Load(); err != nil {
		log.Println("Нет файла .env, используем переменные окружения")
	}

	// Настройка логирования
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetLevel(logrus.InfoLevel)
	logrus.SetOutput(os.Stdout)

	logrus.Info("Запуск банковского сервиса")

	// Подключение к PostgreSQL
	connString := os.Getenv("DB_CONN_STRING")
	if connString == "" {
		connString = "postgres://myuser:mypassword@localhost:5432/bankDB?sslmode=disable"
	}

	db, err := sql.Open("postgres", connString)
	if err != nil {
		logrus.WithError(err).Fatal("Ошибка подключения к БД")
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		logrus.WithError(err).Fatal("Ошибка ping БД")
	}

	logrus.Info("Подключение к PostgreSQL успешно")

	// Инициализация репозиториев
	userRepo := repository.NewUserRepository(db)
	accountRepo := repository.NewAccountRepository(db)
	cardRepo := repository.NewCardRepository(db)
	transactionRepo := repository.NewTransactionRepository(db)
	creditRepo := repository.NewCreditRepository(db)
	scheduleRepo := repository.NewScheduleRepository(db)

	// Инициализация сервисов
	emailService := service.NewEmailService()
	cbrClient := service.NewCBRClient()

	// Исправлено: передаем emailService третьим аргументом
	authService := service.NewAuthService(userRepo, os.Getenv("JWT_SECRET"), emailService)
	accountService := service.NewAccountService(accountRepo, userRepo)
	cardService := service.NewCardService(cardRepo, accountRepo, emailService)
	transferService := service.NewTransferService(accountRepo, transactionRepo, emailService)
	creditService := service.NewCreditService(creditRepo, accountRepo, scheduleRepo, emailService, cbrClient)
	analyticsService := service.NewAnalyticsService(accountRepo, transactionRepo, creditRepo, scheduleRepo, cbrClient)

	// Инициализация хендлеров
	authHandler := handler.NewAuthHandler(authService)
	accountHandler := handler.NewAccountHandler(accountService)
	cardHandler := handler.NewCardHandler(cardService)
	transferHandler := handler.NewTransferHandler(transferService)
	creditHandler := handler.NewCreditHandler(creditService)
	analyticsHandler := handler.NewAnalyticsHandler(analyticsService)
	webHandler := handler.NewWebHandler()

	// Настройка middleware
	authMiddleware := middleware.NewAuthMiddleware(os.Getenv("JWT_SECRET"))

	// Настройка роутера
	r := mux.NewRouter()

	// Статические файлы
	r.PathPrefix("/css/").HandlerFunc(webHandler.ServeStatic)
	r.PathPrefix("/js/").HandlerFunc(webHandler.ServeStatic)
	r.HandleFunc("/dashboard.html", webHandler.ServeDashboard).Methods("GET")

	// Публичные маршруты
	r.HandleFunc("/", webHandler.ServeIndex).Methods("GET")
	r.HandleFunc("/register", authHandler.Register).Methods("POST")
	r.HandleFunc("/login", authHandler.Login).Methods("POST")

	// Защищенные маршруты (с JWT)
	api := r.PathPrefix("/api").Subrouter()
	// Исправлено: оборачиваем handler правильно
	api.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authMiddleware.Authenticate(next.ServeHTTP)(w, r)
		})
	})

	api.HandleFunc("/accounts", accountHandler.CreateAccount).Methods("POST")
	api.HandleFunc("/accounts", accountHandler.GetAccounts).Methods("GET")
	api.HandleFunc("/accounts/{id}/deposit", accountHandler.Deposit).Methods("POST")
	api.HandleFunc("/transfer", transferHandler.Transfer).Methods("POST")
	api.HandleFunc("/cards", cardHandler.CreateCard).Methods("POST")
	api.HandleFunc("/cards/{accountId}", cardHandler.GetCards).Methods("GET")
	api.HandleFunc("/cards/{id}/pay", cardHandler.Pay).Methods("POST")
	api.HandleFunc("/credits", creditHandler.CreateCredit).Methods("POST")
	api.HandleFunc("/credits/{id}/schedule", creditHandler.GetSchedule).Methods("GET")
	api.HandleFunc("/analytics", analyticsHandler.GetAnalytics).Methods("GET")
	api.HandleFunc("/accounts/{id}/predict", analyticsHandler.PredictBalance).Methods("GET")

	// Запуск шедулера для обработки просроченных платежей
	go startScheduler(creditService)

	// Запуск сервера
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	logrus.WithField("port", port).Info("Сервер запущен")
	log.Fatal(http.ListenAndServe(":"+port, r))
}

// Запуск шедулера (каждые 12 часов)
func startScheduler(creditService *service.CreditService) {
	ticker := time.NewTicker(12 * time.Hour)
	for range ticker.C {
		logrus.Info("Запуск шедулера: обработка просроченных платежей")
		creditService.ProcessOverduePayments()
	}
}
