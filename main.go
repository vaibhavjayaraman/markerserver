package main

import (
	"github.com/worldhistorymap/backend/pkg/server"
	"github.com/worldhistorymap/backend/pkg/shared"
	"go.uber.org/zap"
)

func main() {
	logger, err := shared.GetLogger()
	defer logger.Sync()
	if err != nil {
		logger.Fatal("Failed to create Logger", zap.Error(err))
		return
	}

	config, err := shared.GetConfig()
	if err != nil {
		logger.Fatal("Failed to create config", zap.Error(err))
		return
	}

	s, err := markerserver.NewServer(config, logger)
	if err != nil {
		logger.Fatal("Failed to create new marker server", zap.Error(err))
		return
	}

	if err := s.Run(); err != nil {
		logger.Fatal("Error while running marker server", zap.Error(err))
		return
	}
}
