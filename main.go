package main

import (
	"github.com/carson-networks/budget-ingest/api"
	"github.com/carson-networks/budget-ingest/internal/config"
	"github.com/carson-networks/budget-ingest/internal/logging"
)

func main() {
	logger := logging.SetupLogging()
	logger.Info("budget-ingest starting")

	_, err := config.ProcessEnvironmentVariables()
	if err != nil {
		logger.WithError(err).Fatal("config.ProcessEnvironmentVariables")
	}

	srv := api.Server{
		Logger: logger,
		Port:   "9448",
	}
	srv.Serve()
}
