package internal

import (
	"context"
	"github.com/ctfloyd/hazelmere-api/src/internal/database"
	"github.com/ctfloyd/hazelmere-api/src/internal/initialize"
	"github.com/ctfloyd/hazelmere-api/src/internal/snapshot"
	"github.com/ctfloyd/hazelmere-api/src/internal/user"
	"github.com/ctfloyd/hazelmere-commons/pkg/hz_handler"
	"github.com/ctfloyd/hazelmere-commons/pkg/hz_logger"
	"github.com/go-chi/chi/v5"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"os"
)

type Application struct {
	Router      *chi.Mux
	MongoClient *mongo.Client
}

func (app *Application) Init(ctx context.Context, l hz_logger.Logger) {
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
		UserCollectionName:     "user",
	})

	sc := f.NewSnapshotCollection()
	sr := snapshot.NewSnapshotRepository(sc, l)
	sv := snapshot.NewSnapshotValidator()
	ss := snapshot.NewSnapshotService(l, sr, sv)
	sh := snapshot.NewSnapshotHandler(l, ss)

	uc := f.NewUserCollection()
	ur := user.NewUserRepository(uc, l)
	uv := user.NewUserValidator()
	us := user.NewUserService(l, ur, uv)
	uh := user.NewUserHandler(l, us)

	l.Info(context.TODO(), "Init router.")
	handlers := []hz_handler.HazelmereHandler{sh, uh}
	for i := 0; i < len(handlers); i++ {
		handlers[i].RegisterRoutes(app.Router, hz_handler.ApiVersionV1)
	}

	l.Info(context.TODO(), "Done init.")
}

func (app *Application) Cleanup(ctx context.Context) {
	initialize.MongoCleanup(ctx, app.MongoClient)
}
