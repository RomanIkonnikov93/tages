package main

import (
	"github.com/RomanIkonnikov93/tages/internal/config"
	"github.com/RomanIkonnikov93/tages/internal/grpcapi"
	"github.com/RomanIkonnikov93/tages/internal/server"
	"github.com/RomanIkonnikov93/tages/pkg/pkg/logging"
)

func main() {

	logger := logging.GetLogger()

	cfg, err := config.GetConfig()
	if err != nil {
		logger.Fatalf("GetConfig: %s", err)
	}

	service := grpcapi.InitServices(logger)

	go service.Run()

	err = server.StartServer(service, cfg, logger)
	if err != nil {
		logger.Fatalf("StartServer: %s", err)
	}
}
