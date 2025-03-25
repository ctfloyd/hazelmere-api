package internal

import (
	"api/src/internal/common/database"
	"api/src/internal/common/handler"
	"api/src/internal/common/logger"
	"api/src/internal/initialize"
	"api/src/internal/snapshot"
	"context"
	"github.com/go-chi/chi/v5"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"os"
)

type Application struct {
	Router      *chi.Mux
	MongoClient *mongo.Client
}

func (app *Application) Init(ctx context.Context, l logger.Logger) {
	l.Info(context.TODO(), "Init Hazelmere web service.")

	router := initialize.InitRouter(l)
	app.Router = router

	l.Info(context.TODO(), "Init mongo.")

	username := os.Getenv("MONGOUSER")
	password := os.Getenv("MONGOPASSWORD")
	host := os.Getenv("MONGO_URL")

	c, err := initialize.MongoClient(
		host,
		username,
		password,
	)
	if err != nil {
		l.InfoArgs(ctx, "failed to init mc: %v", err)
	}
	app.MongoClient = c

	f := database.NewMongoFactory(c, database.MongoFactoryConfig{
		DatabaseName:           "hazelmere",
		SnapshotCollectionName: "snapshot",
	})
	sc := f.NewSnapshotCollection()
	sr := snapshot.NewSnapshotRepository(sc, l)
	sv := snapshot.NewSnapshotValidator()
	ss := snapshot.NewSnapshotService(l, sr, sv)
	sh := snapshot.NewSnapshotHandler(l, ss)

	l.Info(context.TODO(), "Init router.")
	handlers := []handler.HazelmereHandler{sh}
	for i := 0; i < len(handlers); i++ {
		handlers[i].RegisterRoutes(app.Router, handler.ApiVersionV1)
	}

	l.Info(context.TODO(), "Done init.")
}

func (app *Application) Cleanup(ctx context.Context) {
	initialize.MongoCleanup(ctx, app.MongoClient)
}
