package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestHealthService_GetHealth(t *testing.T) {
	testHelper := serviceTestHelper(t)

	tests := []struct {
		name     string
		doMock   func()
		wantResp map[string]string
	}{
		{
			name: "success case",
			doMock: func() {
				testHelper.mockCacheRepository.EXPECT().Ping(gomock.Any()).Return(nil)
				testHelper.mockMySQLRepository.EXPECT().Ping(gomock.Any()).Return(nil)
			},
			wantResp: map[string]string{
				"redis": "redis is up and running",
				"mysql": "mysql is up and running",
			},
		},
		{
			name: "error case - failed connect to redis",
			doMock: func() {
				testHelper.mockCacheRepository.EXPECT().Ping(gomock.Any()).Return(errors.New("redis error"))
				testHelper.mockMySQLRepository.EXPECT().Ping(gomock.Any()).Return(nil)
			},
			wantResp: map[string]string{
				"redis": "failed connect to redis: redis error",
				"mysql": "mysql is up and running",
			},
		},
		{
			name: "error case - failed connect to mysql",
			doMock: func() {
				testHelper.mockCacheRepository.EXPECT().Ping(gomock.Any()).Return(nil)
				testHelper.mockMySQLRepository.EXPECT().Ping(gomock.Any()).Return(errors.New("mysql error"))
			},
			wantResp: map[string]string{
				"redis": "redis is up and running",
				"mysql": "failed connect to mysql: mysql error",
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.doMock != nil {
				tc.doMock()
			}
			resp := testHelper.healthService.GetHealth(context.TODO())
			assert.Equal(t, tc.wantResp, resp)
		})
	}
}
