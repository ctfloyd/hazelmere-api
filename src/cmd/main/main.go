package main

import (
	"context"
	"github.com/ctfloyd/hazelmere-api/src/internal"
	"github.com/ctfloyd/hazelmere-commons/pkg/hz_logger"
	"net/http"
	"os"
	"os/signal"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	l := hz_logger.NewZeroLogAdapater(hz_logger.LogLevelDebug)

	app := internal.Application{}
	app.Init(ctx, l)

	l.Info(ctx, "Trying listen and serve 8080.")
	err := http.ListenAndServe(":8080", app.Router)
	if err != nil {
		l.InfoArgs(ctx, "Failed to listen and serve on port 8080: %v", err)
	}

	defer app.Cleanup(ctx)
}
