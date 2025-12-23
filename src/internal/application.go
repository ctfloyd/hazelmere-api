package internal

import (
	"context"
	"net/http"

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

type Application struct {
	Router      *chi.Mux
	MongoClient *mongo.Client
}

func (app *Application) Init(logger hz_logger.Logger, config *hz_config.Config) {
	logger.Info(context.TODO(), "Init Hazelmere web service.")

	router := initialize.InitRouter(logger)
	app.Router = router

	logger.Info(context.TODO(), "Init mongo.")

	c, err := initialize.MongoClient(
		config.ValueOrPanic("mongo.connection.host"),
		config.ValueOrPanic("mongo.connection.username"),
		config.ValueOrPanic("mongo.connection.password"),
	)
	if err != nil {
		panic(err)
	}
	app.MongoClient = c

	f := database.NewMongoFactory(c, database.MongoFactoryConfig{
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
	txManager := database.NewTransactionManager(c)
	orchestrator := hiscore.NewHiscoreOrchestrator(logger, snapshotService, deltaService, txManager)

	snapshotHandler := handler.NewSnapshotHandler(logger, snapshotService, orchestrator)

	// Prime delta cache
	logger.Info(context.TODO(), "Starting delta cache priming...")
	if err := deltaService.PrimeCache(context.TODO()); err != nil {
		logger.ErrorArgs(context.TODO(), "Failed to prime delta cache: %v", err)
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

	logger.Info(context.TODO(), "Init router.")
	handlers := []handler.HazelmereHandler{healthHandler, snapshotHandler, userHandler, workerHandler, deltaHandler}
	for i := 0; i < len(handlers); i++ {
		handlers[i].RegisterRoutes(app.Router, handler.ApiVersionV1, authorizer)
	}

	logger.Info(context.TODO(), "Done init.")
}

func (app *Application) Run(ctx context.Context, logger hz_logger.Logger) {
	logger.Info(ctx, "Trying listen and serve 8080.")
	err := http.ListenAndServe(":8080", app.Router)
	if err != nil {
		logger.ErrorArgs(ctx, "Failed to listen and serve on port 8080: %v", err)
	}
}

func (app *Application) Cleanup(ctx context.Context) {
	initialize.MongoCleanup(ctx, app.MongoClient)
}
