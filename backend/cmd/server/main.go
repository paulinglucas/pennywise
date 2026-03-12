package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/jamespsullivan/pennywise/internal/api"
	pennywisecrypto "github.com/jamespsullivan/pennywise/internal/crypto"
	"github.com/jamespsullivan/pennywise/internal/db"
	"github.com/jamespsullivan/pennywise/internal/db/queries"
	"github.com/jamespsullivan/pennywise/internal/dlq"
	"github.com/jamespsullivan/pennywise/internal/middleware"
	"github.com/jamespsullivan/pennywise/internal/observability"
	"github.com/jamespsullivan/pennywise/internal/simplefin"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "migrate" {
		runMigrate()
		return
	}

	runServer()
}

func runServer() {
	logger := newLogger()
	slog.SetDefault(logger)

	shutdownTracer, err := observability.InitTracer(logger, os.Stdout)
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = shutdownTracer(context.Background()) }()

	database := openDatabase()
	defer func() { _ = database.Close() }()

	handler, sfinScheduler := buildRouter(logger, database)
	if sfinScheduler != nil {
		sfinScheduler.Start()
		defer sfinScheduler.Stop()
	}

	port := os.Getenv("PENNYWISE_PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		logger.Info("server starting", slog.String("port", port))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	awaitShutdown(logger, server)
}

func awaitShutdown(logger *slog.Logger, server *http.Server) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("server shutting down")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}

	logger.Info("server stopped")
}

func openDatabase() *sql.DB {
	dbPath := dbPathFromEnv()
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o750); err != nil {
		log.Fatal(err)
	}

	database, err := db.Open(dbPath)
	if err != nil {
		log.Fatal(err)
	}

	if err := db.Migrate(database); err != nil {
		log.Fatal(err)
	}

	return database
}

func buildRouter(logger *slog.Logger, database *sql.DB) (http.Handler, *simplefin.Scheduler) {
	secret := jwtSecret()
	userRepo := queries.NewUserRepository(database)
	accountRepo := queries.NewAccountRepository(database)
	txnRepo := queries.NewTransactionRepository(database)
	groupRepo := queries.NewTransactionGroupRepository(database)
	assetRepo := queries.NewAssetRepository(database)
	goalRepo := queries.NewGoalRepository(database)
	goalContribRepo := queries.NewGoalContributionRepository(database)
	recurringRepo := queries.NewRecurringRepository(database)
	alertRepo := queries.NewAlertRepository(database)
	dashboardRepo := queries.NewDashboardRepository(database)
	auditRepo := queries.NewAuditLogRepository(database)
	dlqWriter := dlq.NewFailedRequestWriter(database)
	handler := api.NewAppHandler(userRepo, accountRepo, txnRepo, groupRepo, assetRepo, goalRepo, goalContribRepo, recurringRepo, alertRepo, dashboardRepo, auditRepo, dlqWriter, secret)

	validator, err := middleware.Validation(api.OpenAPISpec, "/api/v1")
	if err != nil {
		log.Fatal(err)
	}

	apiRouter := chi.NewRouter()
	apiRouter.Use(middleware.RequestID)
	apiRouter.Use(middleware.Logging(logger))
	apiRouter.Use(observability.MetricsMiddleware)
	apiRouter.Use(validator)

	authMiddleware := middleware.Auth(secret, api.CookieAuthScopes)

	apiHandler := api.HandlerWithOptions(handler, api.ChiServerOptions{
		BaseRouter:  apiRouter,
		BaseURL:     "/api/v1",
		Middlewares: []api.MiddlewareFunc{authMiddleware},
	})

	encKey := pennywisecrypto.DeriveKey(string(secret))
	sfinRepo := simplefin.NewSimplefinRepository(database)
	sfinClient := simplefin.NewClient(nil)
	sfinSync := simplefin.NewSyncService(sfinClient, sfinRepo, encKey)
	sfinHandler := simplefin.NewHandler(sfinRepo, sfinClient, sfinSync, encKey)
	sfinRoutes := simplefin.Routes(sfinHandler, secret)

	syncHour := 6
	if h := os.Getenv("PENNYWISE_SYNC_HOUR"); h != "" {
		if parsed, err := strconv.Atoi(h); err == nil && parsed >= 0 && parsed <= 23 {
			syncHour = parsed
		}
	}
	scheduler := simplefin.NewScheduler(sfinSync, syncHour)

	root := chi.NewRouter()
	root.Mount("/", apiHandler)
	root.Mount("/api/v1/simplefin", sfinRoutes)
	root.Handle("/metrics", observability.LocalhostOnly(observability.MetricsHandler()))

	return root, scheduler
}

func runMigrate() {
	dbPath := dbPathFromEnv()

	if err := os.MkdirAll(filepath.Dir(dbPath), 0o750); err != nil {
		log.Fatal(err)
	}

	database, err := db.Open(dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = database.Close() }()

	if err := db.Migrate(database); err != nil {
		log.Fatal(err)
	}

	fmt.Println("migrations applied successfully")
}

func dbPathFromEnv() string {
	path := os.Getenv("PENNYWISE_DB_PATH")
	if path == "" {
		path = "./data/pennywise.db"
	}
	return filepath.Clean(path)
}

func jwtSecret() []byte {
	secret := os.Getenv("PENNYWISE_JWT_SECRET")
	if secret == "" {
		log.Fatal("PENNYWISE_JWT_SECRET environment variable is required")
	}
	return []byte(secret)
}

func newLogger() *slog.Logger {
	level := slog.LevelInfo
	if os.Getenv("PENNYWISE_LOG_LEVEL") == "debug" {
		level = slog.LevelDebug
	}
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	}))
}
