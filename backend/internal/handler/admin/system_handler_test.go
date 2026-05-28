package admin

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type systemHandlerUpdateCacheStub struct{}

func (systemHandlerUpdateCacheStub) GetUpdateInfo(context.Context) (string, error) {
	return "", fmt.Errorf("cache miss")
}

func (systemHandlerUpdateCacheStub) SetUpdateInfo(context.Context, string, time.Duration) error {
	return nil
}

type systemHandlerGitHubClientStub struct {
	release *service.GitHubRelease
}

func (s systemHandlerGitHubClientStub) FetchLatestRelease(context.Context, string) (*service.GitHubRelease, error) {
	return s.release, nil
}

func (systemHandlerGitHubClientStub) DownloadFile(context.Context, string, string, int64) error {
	return nil
}

func (systemHandlerGitHubClientStub) FetchChecksumFile(context.Context, string) ([]byte, error) {
	return nil, nil
}

func TestSystemHandlerPerformUpdateReturnsConflictWhenAlreadyUpToDate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	prevCoordinator := service.DefaultIdempotencyCoordinator()
	service.SetDefaultIdempotencyCoordinator(nil)
	t.Cleanup(func() {
		service.SetDefaultIdempotencyCoordinator(prevCoordinator)
	})

	updateSvc := service.NewUpdateService(
		systemHandlerUpdateCacheStub{},
		systemHandlerGitHubClientStub{
			release: &service.GitHubRelease{
				TagName: "v1.1.0",
			},
		},
		"1.1.0",
		"release",
		config.UpdateConfig{},
	)

	lockRepo := newMemoryIdempotencyRepoStub()
	lockSvc := service.NewSystemOperationLockService(lockRepo, service.DefaultIdempotencyConfig())
	handler := NewSystemHandler(updateSvc, lockSvc)

	router := gin.New()
	router.POST("/api/v1/admin/system/update", handler.PerformUpdate)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/system/update", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusConflict, rec.Code)

	var resp struct {
		Code    int    `json:"code"`
		Reason  string `json:"reason"`
		Message string `json:"message"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Equal(t, http.StatusConflict, resp.Code)
	require.Equal(t, "NO_UPDATE_AVAILABLE", resp.Reason)
	require.Contains(t, resp.Message, "latest version")
}
