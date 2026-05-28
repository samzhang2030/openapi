package service

import (
	"context"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/stretchr/testify/require"
)

type updateServiceCacheStub struct{}

func (updateServiceCacheStub) GetUpdateInfo(context.Context) (string, error) {
	return "", fmt.Errorf("cache miss")
}

func (updateServiceCacheStub) SetUpdateInfo(context.Context, string, time.Duration) error {
	return nil
}

type updateServiceGitHubClientStub struct {
	release       *GitHubRelease
	downloadCalls int
}

type updateServiceExternalRepoStatusStub struct {
	CurrentRef     string
	CurrentCommit  string
	UpstreamRef    string
	UpstreamCommit string
	BehindCount    int
}

func (s *updateServiceGitHubClientStub) FetchLatestRelease(context.Context, string) (*GitHubRelease, error) {
	return s.release, nil
}

func (s *updateServiceGitHubClientStub) DownloadFile(context.Context, string, string, int64) error {
	s.downloadCalls++
	return nil
}

func (s *updateServiceGitHubClientStub) FetchChecksumFile(context.Context, string) ([]byte, error) {
	return nil, nil
}

func TestCheckUpdate_ReleaseBuildMarksAutoUpdateUnsupportedWhenExecutableDirIsNotWritable(t *testing.T) {
	client := &updateServiceGitHubClientStub{
		release: &GitHubRelease{
			TagName: "v1.2.0",
			Assets: []GitHubAsset{
				{
					Name:               fmt.Sprintf("sub2api_%s_%s.tar.gz", runtime.GOOS, runtime.GOARCH),
					BrowserDownloadURL: "https://github.com/Wei-Shaw/sub2api/releases/download/v1.2.0/sub2api.tar.gz",
					Size:               1024,
				},
			},
		},
	}
	service := NewUpdateService(updateServiceCacheStub{}, client, "1.1.0", "release", config.UpdateConfig{})

	prevResolveExecutablePath := resolveUpdateExecutablePath
	prevCreateTempDir := createUpdateTempDir
	prevRemoveAll := removeUpdatePathAll
	prevDetectContainer := detectUpdateContainerRuntime
	resolveUpdateExecutablePath = func() (string, error) {
		return "/app/sub2api", nil
	}
	createUpdateTempDir = func(string, string) (string, error) {
		return "", fmt.Errorf("mkdir /app/.sub2api-update-123: %w", fs.ErrPermission)
	}
	removeUpdatePathAll = func(string) error { return nil }
	detectUpdateContainerRuntime = func() bool { return false }
	t.Cleanup(func() {
		resolveUpdateExecutablePath = prevResolveExecutablePath
		createUpdateTempDir = prevCreateTempDir
		removeUpdatePathAll = prevRemoveAll
		detectUpdateContainerRuntime = prevDetectContainer
	})

	info, err := service.CheckUpdate(context.Background(), true)
	require.NoError(t, err)
	require.True(t, info.HasUpdate)
	require.False(t, info.CanAutoUpdate)
	require.NotEmpty(t, info.UpdateHint)
	require.Contains(t, info.UpdateHint, "online update")
}

func TestCheckUpdate_ContainerReleaseWithoutExternalUpdaterIsUnsupported(t *testing.T) {
	client := &updateServiceGitHubClientStub{
		release: &GitHubRelease{
			TagName: "v1.2.0",
			Assets: []GitHubAsset{
				{
					Name:               fmt.Sprintf("sub2api_%s_%s.tar.gz", runtime.GOOS, runtime.GOARCH),
					BrowserDownloadURL: "https://github.com/Wei-Shaw/sub2api/releases/download/v1.2.0/sub2api.tar.gz",
					Size:               1024,
				},
			},
		},
	}
	service := NewUpdateService(updateServiceCacheStub{}, client, "1.1.0", "release", config.UpdateConfig{})

	prevDetectContainer := detectUpdateContainerRuntime
	detectUpdateContainerRuntime = func() bool { return true }
	t.Cleanup(func() {
		detectUpdateContainerRuntime = prevDetectContainer
	})

	info, err := service.CheckUpdate(context.Background(), true)
	require.NoError(t, err)
	require.True(t, info.HasUpdate)
	require.False(t, info.CanAutoUpdate)
	require.Contains(t, info.UpdateHint, "online update")
}

func TestPerformUpdate_ReturnsConflictWhenAutoUpdateUnsupported(t *testing.T) {
	client := &updateServiceGitHubClientStub{
		release: &GitHubRelease{
			TagName: "v1.2.0",
			Assets: []GitHubAsset{
				{
					Name:               fmt.Sprintf("sub2api_%s_%s.tar.gz", runtime.GOOS, runtime.GOARCH),
					BrowserDownloadURL: "https://github.com/Wei-Shaw/sub2api/releases/download/v1.2.0/sub2api.tar.gz",
					Size:               1024,
				},
			},
		},
	}
	service := NewUpdateService(updateServiceCacheStub{}, client, "1.1.0", "release", config.UpdateConfig{})

	prevResolveExecutablePath := resolveUpdateExecutablePath
	prevCreateTempDir := createUpdateTempDir
	prevRemoveAll := removeUpdatePathAll
	prevDetectContainer := detectUpdateContainerRuntime
	resolveUpdateExecutablePath = func() (string, error) {
		return "/app/sub2api", nil
	}
	createUpdateTempDir = func(string, string) (string, error) {
		return "", fmt.Errorf("mkdir /app/.sub2api-update-123: %w", fs.ErrPermission)
	}
	removeUpdatePathAll = func(string) error { return nil }
	detectUpdateContainerRuntime = func() bool { return false }
	t.Cleanup(func() {
		resolveUpdateExecutablePath = prevResolveExecutablePath
		createUpdateTempDir = prevCreateTempDir
		removeUpdatePathAll = prevRemoveAll
		detectUpdateContainerRuntime = prevDetectContainer
	})

	_, err := service.PerformUpdate(context.Background())
	require.Error(t, err)
	require.Equal(t, http.StatusConflict, infraerrors.Code(err))
	require.Equal(t, "AUTO_UPDATE_UNSUPPORTED", infraerrors.Reason(err))
	require.Contains(t, infraerrors.Message(err), "online update")
	require.Zero(t, client.downloadCalls)
}

func TestCheckUpdate_ExternalUpdaterMarksAutoUpdateSupported(t *testing.T) {
	cmdDir := t.TempDir()
	cmdPath := cmdDir + "/update-helper.sh"
	require.NoError(t, os.WriteFile(cmdPath, []byte("#!/bin/sh\nexit 0\n"), 0o755))

	workDir := t.TempDir()
	client := &updateServiceGitHubClientStub{
		release: &GitHubRelease{
			TagName: "v1.2.0",
			Assets: []GitHubAsset{
				{
					Name:               fmt.Sprintf("sub2api_%s_%s.tar.gz", runtime.GOOS, runtime.GOARCH),
					BrowserDownloadURL: "https://github.com/Wei-Shaw/sub2api/releases/download/v1.2.0/sub2api.tar.gz",
					Size:               1024,
				},
			},
		},
	}
	service := NewUpdateService(updateServiceCacheStub{}, client, "1.1.0", "release", config.UpdateConfig{
		ExternalUpdaterCommand:              cmdPath,
		ExternalUpdaterWorkingDirectory:     workDir,
		ExternalUpdaterHelperTimeoutSeconds: 600,
	})

	prevResolveImage := resolveUpdateHelperImage
	prevRunDockerCheck := runUpdateDockerAccessCheck
	resolveUpdateHelperImage = func(context.Context) (string, error) {
		return "openapi-prod:test", nil
	}
	runUpdateDockerAccessCheck = func(context.Context) error { return nil }
	t.Cleanup(func() {
		resolveUpdateHelperImage = prevResolveImage
		runUpdateDockerAccessCheck = prevRunDockerCheck
	})

	info, err := service.CheckUpdate(context.Background(), true)
	require.NoError(t, err)
	require.True(t, info.HasUpdate)
	require.True(t, info.CanAutoUpdate)
	require.Empty(t, info.UpdateHint)
}

func TestPerformUpdate_StartsDetachedExternalUpdaterHelper(t *testing.T) {
	cmdDir := t.TempDir()
	cmdPath := cmdDir + "/update-helper.sh"
	require.NoError(t, os.WriteFile(cmdPath, []byte("#!/bin/sh\nexit 0\n"), 0o755))

	workDir := t.TempDir()
	client := &updateServiceGitHubClientStub{
		release: &GitHubRelease{
			TagName: "v1.2.0",
			Assets: []GitHubAsset{
				{
					Name:               fmt.Sprintf("sub2api_%s_%s.tar.gz", runtime.GOOS, runtime.GOARCH),
					BrowserDownloadURL: "https://github.com/Wei-Shaw/sub2api/releases/download/v1.2.0/sub2api.tar.gz",
					Size:               1024,
				},
			},
		},
	}
	service := NewUpdateService(updateServiceCacheStub{}, client, "1.1.0", "release", config.UpdateConfig{
		ExternalUpdaterCommand:              cmdPath,
		ExternalUpdaterWorkingDirectory:     workDir,
		ExternalUpdaterHelperTimeoutSeconds: 900,
	})

	prevResolveImage := resolveUpdateHelperImage
	prevRunDockerCheck := runUpdateDockerAccessCheck
	prevLaunchHelper := launchDetachedExternalUpdater
	resolveUpdateHelperImage = func(context.Context) (string, error) {
		return "openapi-prod:test", nil
	}
	runUpdateDockerAccessCheck = func(context.Context) error { return nil }

	var launched updateExternalHelperSpec
	launchDetachedExternalUpdater = func(_ context.Context, spec updateExternalHelperSpec) error {
		launched = spec
		return nil
	}
	t.Cleanup(func() {
		resolveUpdateHelperImage = prevResolveImage
		runUpdateDockerAccessCheck = prevRunDockerCheck
		launchDetachedExternalUpdater = prevLaunchHelper
	})

	result, err := service.PerformUpdate(context.Background())
	require.NoError(t, err)
	require.Equal(t, "openapi-prod:test", launched.HelperImage)
	require.Equal(t, cmdPath, launched.CommandPath)
	require.Equal(t, workDir, launched.WorkDir)
	require.Equal(t, 15*time.Minute, launched.Timeout)
	require.False(t, result.NeedRestart)
	require.True(t, result.PollForRestart)
	require.Contains(t, result.Message, "automatically")
	require.Zero(t, client.downloadCalls)
}

func TestCheckUpdate_ExternalUpdaterUsesRepoStatusWhenWorktreeIsAlreadyCurrent(t *testing.T) {
	cmdDir := t.TempDir()
	cmdPath := cmdDir + "/update-helper.sh"
	require.NoError(t, os.WriteFile(cmdPath, []byte("#!/bin/sh\nexit 0\n"), 0o755))

	workDir := t.TempDir()
	client := &updateServiceGitHubClientStub{
		release: &GitHubRelease{
			TagName: "v1.2.0",
			Assets: []GitHubAsset{
				{
					Name:               fmt.Sprintf("sub2api_%s_%s.tar.gz", runtime.GOOS, runtime.GOARCH),
					BrowserDownloadURL: "https://github.com/Wei-Shaw/sub2api/releases/download/v1.2.0/sub2api.tar.gz",
					Size:               1024,
				},
			},
		},
	}
	service := NewUpdateService(updateServiceCacheStub{}, client, "1.1.0", "release", config.UpdateConfig{
		ExternalUpdaterCommand:              cmdPath,
		ExternalUpdaterWorkingDirectory:     workDir,
		ExternalUpdaterHelperTimeoutSeconds: 600,
	})

	prevResolveImage := resolveUpdateHelperImage
	prevRunDockerCheck := runUpdateDockerAccessCheck
	prevInspectRepoStatus := inspectExternalUpdaterRepoStatus
	resolveUpdateHelperImage = func(context.Context) (string, error) {
		return "openapi-prod:test", nil
	}
	runUpdateDockerAccessCheck = func(context.Context) error { return nil }
	inspectExternalUpdaterRepoStatus = func(context.Context, string) (*externalUpdaterRepoStatus, error) {
		return &externalUpdaterRepoStatus{
			CurrentRef:     "main",
			CurrentCommit:  "1111111111111111111111111111111111111111",
			UpstreamRef:    "origin/main",
			UpstreamCommit: "1111111111111111111111111111111111111111",
			BehindCount:    0,
		}, nil
	}
	t.Cleanup(func() {
		resolveUpdateHelperImage = prevResolveImage
		runUpdateDockerAccessCheck = prevRunDockerCheck
		inspectExternalUpdaterRepoStatus = prevInspectRepoStatus
	})

	info, err := service.CheckUpdate(context.Background(), true)
	require.NoError(t, err)
	require.False(t, info.HasUpdate)
	require.Equal(t, "1.1.0", info.LatestVersion)
}

func TestCheckUpdate_ExternalUpdaterFallsBackToGitRefWhenRepoBehindWithoutNewReleaseVersion(t *testing.T) {
	cmdDir := t.TempDir()
	cmdPath := cmdDir + "/update-helper.sh"
	require.NoError(t, os.WriteFile(cmdPath, []byte("#!/bin/sh\nexit 0\n"), 0o755))

	workDir := t.TempDir()
	client := &updateServiceGitHubClientStub{
		release: &GitHubRelease{
			TagName: "v1.1.0",
			Assets: []GitHubAsset{
				{
					Name:               fmt.Sprintf("sub2api_%s_%s.tar.gz", runtime.GOOS, runtime.GOARCH),
					BrowserDownloadURL: "https://github.com/Wei-Shaw/sub2api/releases/download/v1.1.0/sub2api.tar.gz",
					Size:               1024,
				},
			},
		},
	}
	service := NewUpdateService(updateServiceCacheStub{}, client, "1.1.0", "release", config.UpdateConfig{
		ExternalUpdaterCommand:              cmdPath,
		ExternalUpdaterWorkingDirectory:     workDir,
		ExternalUpdaterHelperTimeoutSeconds: 600,
	})

	prevResolveImage := resolveUpdateHelperImage
	prevRunDockerCheck := runUpdateDockerAccessCheck
	prevInspectRepoStatus := inspectExternalUpdaterRepoStatus
	resolveUpdateHelperImage = func(context.Context) (string, error) {
		return "openapi-prod:test", nil
	}
	runUpdateDockerAccessCheck = func(context.Context) error { return nil }
	inspectExternalUpdaterRepoStatus = func(context.Context, string) (*externalUpdaterRepoStatus, error) {
		return &externalUpdaterRepoStatus{
			CurrentRef:     "main",
			CurrentCommit:  "1111111111111111111111111111111111111111",
			UpstreamRef:    "origin/main",
			UpstreamCommit: "95ecf821f0660d2039572255afb3c389a469820d",
			BehindCount:    1,
		}, nil
	}
	t.Cleanup(func() {
		resolveUpdateHelperImage = prevResolveImage
		runUpdateDockerAccessCheck = prevRunDockerCheck
		inspectExternalUpdaterRepoStatus = prevInspectRepoStatus
	})

	info, err := service.CheckUpdate(context.Background(), true)
	require.NoError(t, err)
	require.True(t, info.HasUpdate)
	require.Equal(t, "origin/main@95ecf821", info.LatestVersion)
}

func TestPerformUpdate_ReturnsConflictWhenAlreadyUpToDate(t *testing.T) {
	client := &updateServiceGitHubClientStub{
		release: &GitHubRelease{
			TagName: "v1.1.0",
			Assets: []GitHubAsset{
				{
					Name:               fmt.Sprintf("sub2api_%s_%s.tar.gz", runtime.GOOS, runtime.GOARCH),
					BrowserDownloadURL: "https://github.com/Wei-Shaw/sub2api/releases/download/v1.1.0/sub2api.tar.gz",
					Size:               1024,
				},
			},
		},
	}
	service := NewUpdateService(updateServiceCacheStub{}, client, "1.1.0", "release", config.UpdateConfig{})

	_, err := service.PerformUpdate(context.Background())
	require.Error(t, err)
	require.Equal(t, http.StatusConflict, infraerrors.Code(err))
	require.Equal(t, "NO_UPDATE_AVAILABLE", infraerrors.Reason(err))
	require.Contains(t, infraerrors.Message(err), "latest version")
	require.Zero(t, client.downloadCalls)
}
