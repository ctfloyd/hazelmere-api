package internal

import (
	"context"
	"github.com/ctfloyd/hazelmere-api/src/internal/common/handler"
	"github.com/ctfloyd/hazelmere-api/src/internal/common/health"
	"github.com/ctfloyd/hazelmere-api/src/internal/database"
	"github.com/ctfloyd/hazelmere-api/src/internal/initialize"
	"github.com/ctfloyd/hazelmere-api/src/internal/middleware"
	"github.com/ctfloyd/hazelmere-api/src/internal/snapshot"
	"github.com/ctfloyd/hazelmere-api/src/internal/user"
	"github.com/ctfloyd/hazelmere-api/src/internal/worker"
	"github.com/ctfloyd/hazelmere-commons/pkg/hz_config"
	"github.com/ctfloyd/hazelmere-commons/pkg/hz_logger"
	"github.com/go-chi/chi/v5"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"net/http"
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
	})

	userCollection := f.NewUserCollection()
	userRepo := user.NewUserRepository(userCollection, logger)
	userValidator := user.NewUserValidator()
	userService := user.NewUserService(logger, userRepo, userValidator)
	userHandler := user.NewUserHandler(logger, userService)

	snapshotCollection := f.NewSnapshotCollection()
	snapshotRepo := snapshot.NewSnapshotRepository(snapshotCollection, logger)
	snapshotValidator := snapshot.NewSnapshotValidator()
	snapshotCache := snapshot.NewSnapshotCache()
	snapshotService := snapshot.NewSnapshotService(logger, snapshotRepo, snapshotValidator, snapshotCache, userRepo)
	snapshotHandler := snapshot.NewSnapshotHandler(logger, snapshotService)

	logger.Info(context.TODO(), "Starting cache priming...")
	if err := snapshotService.PrimeCache(context.TODO()); err != nil {
		logger.ErrorArgs(context.TODO(), "Failed to prime cache: %v", err)
	}

	workerClient := initialize.InitWorkerClient(logger, config)
	workerService := worker.NewWorkerService(logger, workerClient, snapshotService)
	workerHandler := worker.NewWorkerHandler(logger, workerService)

	healthHandler := health.NewHealthHandler(logger)

	authorizer := middleware.NewAuthorizer(
		config.BoolValueOrPanic("auth.enabled"),
		config.StringSliceValueOrPanic("auth.tokens"),
		logger,
	)

	logger.Info(context.TODO(), "Init router.")
	handlers := []handler.HazelmereHandler{healthHandler, snapshotHandler, userHandler, workerHandler}
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
