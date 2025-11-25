package services

import (
	"context"
	"fmt"
)

type HealthService interface {
	GetHealth(ctx context.Context) map[string]string
}

type healthService service

var _ HealthService = (*healthService)(nil)

func (hs *healthService) GetHealth(ctx context.Context) map[string]string {
	status := map[string]string{
		"redis": "redis is up and running",
		"mysql": "mysql is up and running",
	}

	if err := hs.srv.cacheRepo.Ping(ctx); err != nil {
		status["redis"] = fmt.Sprintf("failed connect to redis: %v", err)
	}

	if err := hs.srv.mySqlRepo.Ping(ctx); err != nil {
		status["mysql"] = fmt.Sprintf("failed connect to mysql: %v", err)
	}

	return status
}
