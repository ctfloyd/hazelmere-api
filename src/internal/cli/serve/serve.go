package serve

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"

	"github.com/ctfloyd/hazelmere-api/src/internal/core/delta"
	"github.com/ctfloyd/hazelmere-api/src/internal/core/hiscore"
	"github.com/ctfloyd/hazelmere-api/src/internal/core/snapshot"
	"github.com/ctfloyd/hazelmere-api/src/internal/core/user"
	"github.com/ctfloyd/hazelmere-api/src/internal/core/worker"
	"github.com/ctfloyd/hazelmere-api/src/internal/database"
	"github.com/ctfloyd/hazelmere-api/src/internal/initialize"
	"github.com/ctfloyd/hazelmere-api/src/internal/rest/handler"
	"github.com/ctfloyd/hazelmere-api/src/internal/rest/middleware"
	"github.com/ctfloyd/hazelmere-commons/pkg/hz_config"
	"github.com/ctfloyd/hazelmere-commons/pkg/hz_logger"
	"github.com/go-chi/chi/v5"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func Run(configPath string, args []string) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	config := hz_config.NewConfigFromPath(configPath)
	if err := config.Read(); err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}

	logger := hz_logger.NewZeroLogAdapater(hz_logger.LogLevelFromString(config.ValueOrPanic("log.level")))
	logger.Info(ctx, "Starting Hazelmere API server")

	router := initialize.InitRouter(logger)

	logger.Info(ctx, "Connecting to MongoDB")
	client, err := initialize.MongoClient(
		config.ValueOrPanic("mongo.connection.host"),
		config.ValueOrPanic("mongo.connection.username"),
		config.ValueOrPanic("mongo.connection.password"),
	)
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}
	defer initialize.MongoCleanup(ctx, client)

	if err := initializeApp(ctx, logger, config, router, client); err != nil {
		return err
	}

	logger.Info(ctx, "Listening on :8080")
	if err := http.ListenAndServe(":8080", router); err != nil {
		return fmt.Errorf("server error: %w", err)
	}

	return nil
}

func initializeApp(ctx context.Context, logger hz_logger.Logger, config *hz_config.Config, router *chi.Mux, client *mongo.Client) error {
	f := database.NewMongoFactory(client, database.MongoFactoryConfig{
		DatabaseName:           config.ValueOrPanic("mongo.database.name"),
		SnapshotCollectionName: config.ValueOrPanic("mongo.database.collections.snapshot"),
		UserCollectionName:     config.ValueOrPanic("mongo.database.collections.user"),
		DeltaCollectionName:    config.ValueOrPanic("mongo.database.collections.delta"),
	})

	userCollection := f.NewUserCollection()
	userRepo := user.NewUserRepository(userCollection, logger)
	userValidator := user.NewUserValidator()
	userService := user.NewUserService(logger, userRepo, userValidator)
	userHandler := handler.NewUserHandler(logger, userService)

	// Initialize delta components with cache
	deltaCollection := f.NewDeltaCollection()
	deltaRepo := delta.NewDeltaRepository(deltaCollection, logger)
	deltaCache := delta.NewDeltaCache()
	deltaService := delta.NewDeltaService(logger, deltaRepo, deltaCache, userRepo)
	deltaHandler := handler.NewDeltaHandler(logger, deltaService)

	// Initialize snapshot components
	snapshotCollection := f.NewSnapshotCollection()
	snapshotRepo := snapshot.NewSnapshotRepository(snapshotCollection, logger)
	snapshotValidator := snapshot.NewSnapshotValidator()
	snapshotService := snapshot.NewSnapshotService(logger, snapshotRepo, snapshotValidator, userRepo)

	// Initialize orchestrator (coordinates snapshot and delta creation in transactions)
	txManager := database.NewTransactionManager(client, false)
	orchestrator := hiscore.NewHiscoreOrchestrator(logger, snapshotService, deltaService, txManager)

	snapshotHandler := handler.NewSnapshotHandler(logger, snapshotService, orchestrator)

	// Prime delta cache
	logger.Info(ctx, "Priming delta cache...")
	if err := deltaService.PrimeCache(ctx); err != nil {
		logger.ErrorArgs(ctx, "Failed to prime delta cache: %v", err)
	}

	workerClient := initialize.InitWorkerClient(logger, config)
	workerService := worker.NewWorkerService(logger, workerClient, snapshotService)
	workerHandler := handler.NewWorkerHandler(logger, workerService)

	healthHandler := handler.NewHealthHandler(logger)

	authorizer := middleware.NewAuthorizer(
		config.BoolValueOrPanic("auth.enabled"),
		config.StringSliceValueOrPanic("auth.tokens"),
		logger,
	)

	logger.Info(ctx, "Registering routes")
	handlers := []handler.HazelmereHandler{healthHandler, snapshotHandler, userHandler, workerHandler, deltaHandler}
	for i := 0; i < len(handlers); i++ {
		handlers[i].RegisterRoutes(router, handler.ApiVersionV1, authorizer)
	}

	return nil
}
