package main

import (
	"sa/config"
	"sa/internal/app/server"
	"sa/internal/app/store"
	"sa/internal/pkg/logger"

	"github.com/sirupsen/logrus"
)

func main() {
	conf, err := config.NewConfig("config/config.yaml")
	if err != nil {
		logrus.Fatalf("could not parse yaml config %s\n", err)
	}

	logger := logger.NewLogger(conf.Server.LogLevel)
	store := store.NewConnection(logger)

	server := server.NewServer(conf.Server.Port, logger, store)
	server.Start()
}
