package initialize

import (
	"context"
	"github.com/ctfloyd/hazelmere-commons/pkg/hz_client"
	"github.com/ctfloyd/hazelmere-commons/pkg/hz_config"
	"github.com/ctfloyd/hazelmere-commons/pkg/hz_logger"
	"github.com/ctfloyd/hazelmere-worker/src/pkg/worker_client"
)

func InitWorkerClient(logger hz_logger.Logger, config *hz_config.Config) *worker_client.HazelmereWorker {
	clientConfig := hz_client.HttpClientConfig{
		Host:           config.ValueOrPanic("clients.worker.host"),
		TimeoutMs:      config.IntValueOrPanic("clients.worker.timeout"),
		Retries:        config.IntValueOrPanic("clients.worker.retries"),
		RetryWaitMs:    config.IntValueOrPanic("clients.worker.retryWaitMs"),
		RetryMaxWaitMs: config.IntValueOrPanic("clients.worker.retryMaxWaitMs"),
	}
	httpClient := hz_client.NewHttpClient(clientConfig, func(msg string) { logger.Error(context.TODO(), msg) })
	return worker_client.NewHazelmereWorker(httpClient)
}
