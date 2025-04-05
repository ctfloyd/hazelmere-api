package internal

import (
	"context"
	"github.com/ctfloyd/hazelmere-api/src/internal/database"
	"github.com/ctfloyd/hazelmere-api/src/internal/initialize"
	"github.com/ctfloyd/hazelmere-api/src/internal/snapshot"
	"github.com/ctfloyd/hazelmere-api/src/internal/user"
	"github.com/ctfloyd/hazelmere-commons/pkg/hz_config"
	"github.com/ctfloyd/hazelmere-commons/pkg/hz_handler"
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

	snapshotCollection := f.NewSnapshotCollection()
	snapshotRepo := snapshot.NewSnapshotRepository(snapshotCollection, logger)
	snapshotValidator := snapshot.NewSnapshotValidator()
	snapshotService := snapshot.NewSnapshotService(logger, snapshotRepo, snapshotValidator)
	snapshotHandler := snapshot.NewSnapshotHandler(logger, snapshotService)

	userCollection := f.NewUserCollection()
	userRepo := user.NewUserRepository(userCollection, logger)
	userValidator := user.NewUserValidator()
	userService := user.NewUserService(logger, userRepo, userValidator)
	userHandler := user.NewUserHandler(logger, userService)

	logger.Info(context.TODO(), "Init router.")
	handlers := []hz_handler.HazelmereHandler{snapshotHandler, userHandler}
	for i := 0; i < len(handlers); i++ {
		handlers[i].RegisterRoutes(app.Router, hz_handler.ApiVersionV1)
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
