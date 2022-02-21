package main

import (
	"sa/config"
	"sa/internal/app/server"
	"sa/internal/pkg/logger"

	"github.com/sirupsen/logrus"
)

func main() {
	conf, err := config.NewConfig("config/config.yaml")
	if err != nil {
		logrus.Fatalf("could not parse yaml config %s\n", err)
	}

	logger := logger.NewLogger(conf.Server.LogLevel)

	server := server.NewServer(conf.Server.Port, logger)
	server.Start()
}
