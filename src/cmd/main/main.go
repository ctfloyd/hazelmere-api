package main

import (
	"api/src/internal"
	"api/src/internal/common/logger"
	"context"
	"net/http"
	"os"
	"os/signal"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	l := logger.NewZeroLogAdapater(logger.LogLevelDebug)

	app := internal.Application{}
	app.Init(ctx, l)

	l.Info(ctx, "Trying listen and serve 8080.")
	err := http.ListenAndServe(":8080", app.Router)
	if err != nil {
		l.InfoArgs(ctx, "Failed to listen and serve on port 8080: %v", err)
	}

	defer app.Cleanup(ctx)
}
