package service

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

const (
	updateCacheKey = "update_check_cache"
	updateCacheTTL = 1200 // 20 minutes
	githubRepo     = "Wei-Shaw/sub2api"

	// Security: allowed download domains for updates
	allowedDownloadHost = "github.com"
	allowedAssetHost    = "objects.githubusercontent.com"

	// Security: max download size (500MB)
	maxDownloadSize = 500 * 1024 * 1024

	autoUpdateUnsupportedReason = "AUTO_UPDATE_UNSUPPORTED"
	autoUpdateCheckFailedReason = "AUTO_UPDATE_CHECK_FAILED"
	sourceBuildUpdateHint       = "当前为源码构建，请使用 git pull 更新后重启服务 (source build detected; use git pull to update, then restart the service)"
	manualUpdateRequiredHint    = "当前部署环境不支持在线更新，请拉取新镜像或手动替换二进制后重启服务 (online update is not supported in this deployment; pull a new image or replace the binary manually, then restart the service)"
	externalUpdateStartedMsg    = "更新已启动，服务将在新镜像准备完成后自动重启 (update started; the service will restart automatically when the new image is ready)"
)

var (
	resolveUpdateExecutablePath = func() (string, error) {
		exePath, err := os.Executable()
		if err != nil {
			return "", err
		}
		return filepath.EvalSymlinks(exePath)
	}
	createUpdateTempDir = os.MkdirTemp
	removeUpdatePathAll = os.RemoveAll
	resolveUpdateHelperImage = defaultResolveUpdateHelperImage
	runUpdateDockerAccessCheck = defaultRunUpdateDockerAccessCheck
	launchDetachedExternalUpdater = defaultLaunchDetachedExternalUpdater
	inspectExternalUpdaterRepoStatus = defaultInspectExternalUpdaterRepoStatus
	ErrNoUpdateAvailable = infraerrors.Conflict(
		"NO_UPDATE_AVAILABLE",
		"already running the latest version",
	)
)

// UpdateCache defines cache operations for update service
type UpdateCache interface {
	GetUpdateInfo(ctx context.Context) (string, error)
	SetUpdateInfo(ctx context.Context, data string, ttl time.Duration) error
}

// GitHubReleaseClient 获取 GitHub release 信息的接口
type GitHubReleaseClient interface {
	FetchLatestRelease(ctx context.Context, repo string) (*GitHubRelease, error)
	DownloadFile(ctx context.Context, url, dest string, maxSize int64) error
	FetchChecksumFile(ctx context.Context, url string) ([]byte, error)
}

// UpdateService handles software updates
type UpdateService struct {
	cache          UpdateCache
	githubClient   GitHubReleaseClient
	currentVersion string
	buildType      string // "source" for manual builds, "release" for CI builds
	updateConfig   config.UpdateConfig
}

// NewUpdateService creates a new UpdateService
func NewUpdateService(cache UpdateCache, githubClient GitHubReleaseClient, version, buildType string, updateConfig config.UpdateConfig) *UpdateService {
	return &UpdateService{
		cache:          cache,
		githubClient:   githubClient,
		currentVersion: version,
		buildType:      buildType,
		updateConfig:   updateConfig,
	}
}

// UpdateInfo contains update information
type UpdateInfo struct {
	CurrentVersion string       `json:"current_version"`
	LatestVersion  string       `json:"latest_version"`
	HasUpdate      bool         `json:"has_update"`
	ReleaseInfo    *ReleaseInfo `json:"release_info,omitempty"`
	Cached         bool         `json:"cached"`
	Warning        string       `json:"warning,omitempty"`
	BuildType      string       `json:"build_type"` // "source" or "release"
	CanAutoUpdate  bool         `json:"can_auto_update"`
	UpdateHint     string       `json:"update_hint,omitempty"`
}

type UpdateExecutionResult struct {
	Message        string `json:"message"`
	NeedRestart    bool   `json:"need_restart"`
	PollForRestart bool   `json:"poll_for_restart,omitempty"`
}

type updateExternalHelperSpec struct {
	HelperImage string
	CommandPath string
	WorkDir     string
	Timeout     time.Duration
}

type externalUpdaterRepoStatus struct {
	CurrentRef     string
	CurrentCommit  string
	UpstreamRef    string
	UpstreamCommit string
	BehindCount    int
}

// ReleaseInfo contains GitHub release details
type ReleaseInfo struct {
	Name        string  `json:"name"`
	Body        string  `json:"body"`
	PublishedAt string  `json:"published_at"`
	HTMLURL     string  `json:"html_url"`
	Assets      []Asset `json:"assets,omitempty"`
}

// Asset represents a release asset
type Asset struct {
	Name        string `json:"name"`
	DownloadURL string `json:"download_url"`
	Size        int64  `json:"size"`
}

// GitHubRelease represents GitHub API response
type GitHubRelease struct {
	TagName     string        `json:"tag_name"`
	Name        string        `json:"name"`
	Body        string        `json:"body"`
	PublishedAt string        `json:"published_at"`
	HTMLURL     string        `json:"html_url"`
	Assets      []GitHubAsset `json:"assets"`
}

type GitHubAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}

// CheckUpdate checks for available updates
func (s *UpdateService) CheckUpdate(ctx context.Context, force bool) (*UpdateInfo, error) {
	// Try cache first
	if !force {
		if cached, err := s.getFromCache(ctx); err == nil && cached != nil {
			s.applyExternalUpdaterRepoStatus(ctx, cached)
			s.applyAutoUpdateSupport(cached)
			return cached, nil
		}
	}

	// Fetch from GitHub
	info, err := s.fetchLatestRelease(ctx)
	if err != nil {
		// Return cached on error
		if cached, cacheErr := s.getFromCache(ctx); cacheErr == nil && cached != nil {
			cached.Warning = "Using cached data: " + err.Error()
			s.applyAutoUpdateSupport(cached)
			return cached, nil
		}
		info := &UpdateInfo{
			CurrentVersion: s.currentVersion,
			LatestVersion:  s.currentVersion,
			HasUpdate:      false,
			Warning:        err.Error(),
			BuildType:      s.buildType,
		}
		s.applyAutoUpdateSupport(info)
		return info, nil
	}

	// Cache result
	s.saveToCache(ctx, info)
	s.applyExternalUpdaterRepoStatus(ctx, info)
	s.applyAutoUpdateSupport(info)
	return info, nil
}

// PerformUpdate downloads and applies the update
// Uses atomic file replacement pattern for safe in-place updates
func (s *UpdateService) PerformUpdate(ctx context.Context) (*UpdateExecutionResult, error) {
	info, err := s.CheckUpdate(ctx, true)
	if err != nil {
		return nil, err
	}

	if !info.HasUpdate {
		return nil, ErrNoUpdateAvailable.WithMetadata(map[string]string{
			"current_version": info.CurrentVersion,
			"latest_version":  info.LatestVersion,
		})
	}

	if helperSpec, supported, supportHint, err := s.resolveExternalUpdater(ctx); err != nil {
		return nil, err
	} else if supported {
		if err := launchDetachedExternalUpdater(ctx, helperSpec); err != nil {
			return nil, infraerrors.InternalServer(
				autoUpdateCheckFailedReason,
				"无法启动在线更新任务，请查看服务日志 (failed to start online update helper)",
			).WithCause(err)
		}
		return &UpdateExecutionResult{
			Message:        externalUpdateStartedMsg,
			NeedRestart:    false,
			PollForRestart: true,
		}, nil
	} else if supportHint != "" && strings.TrimSpace(s.updateConfig.ExternalUpdaterCommand) != "" {
		return nil, infraerrors.Conflict(autoUpdateUnsupportedReason, supportHint)
	}

	// Find matching archive and checksum for current platform
	archiveName := s.getArchiveName()
	var downloadURL string
	var checksumURL string

	for _, asset := range info.ReleaseInfo.Assets {
		if strings.Contains(asset.Name, archiveName) && !strings.HasSuffix(asset.Name, ".txt") {
			downloadURL = asset.DownloadURL
		}
		if asset.Name == "checksums.txt" {
			checksumURL = asset.DownloadURL
		}
	}

	if downloadURL == "" {
		return nil, fmt.Errorf("no compatible release found for %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	// SECURITY: Validate download URL is from trusted domain
	if err := validateDownloadURL(downloadURL); err != nil {
		return nil, fmt.Errorf("invalid download URL: %w", err)
	}
	if checksumURL != "" {
		if err := validateDownloadURL(checksumURL); err != nil {
			return nil, fmt.Errorf("invalid checksum URL: %w", err)
		}
	}

	if err := s.ensureAutoUpdateSupported(ctx); err != nil {
		return nil, err
	}

	// Get current executable path
	exePath, err := resolveUpdateExecutablePath()
	if err != nil {
		return nil, fmt.Errorf("failed to resolve executable path: %w", err)
	}

	exeDir := filepath.Dir(exePath)

	// Create temp directory in the SAME directory as executable
	// This ensures os.Rename is atomic (same filesystem)
	tempDir, err := createUpdateTempDir(exeDir, ".sub2api-update-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer func() { _ = removeUpdatePathAll(tempDir) }()

	// Download archive
	archivePath := filepath.Join(tempDir, filepath.Base(downloadURL))
	if err := s.downloadFile(ctx, downloadURL, archivePath); err != nil {
		return nil, fmt.Errorf("download failed: %w", err)
	}

	// Verify checksum if available
	if checksumURL != "" {
		if err := s.verifyChecksum(ctx, archivePath, checksumURL); err != nil {
			return nil, fmt.Errorf("checksum verification failed: %w", err)
		}
	}

	// Extract binary from archive
	newBinaryPath := filepath.Join(tempDir, "sub2api")
	if err := s.extractBinary(archivePath, newBinaryPath); err != nil {
		return nil, fmt.Errorf("extraction failed: %w", err)
	}

	// Set executable permission before replacement
	if err := os.Chmod(newBinaryPath, 0755); err != nil {
		return nil, fmt.Errorf("chmod failed: %w", err)
	}

	// Atomic replacement using rename pattern:
	// 1. Rename current -> backup (atomic on Unix)
	// 2. Rename new -> current (atomic on Unix, same filesystem)
	// If step 2 fails, restore backup
	backupPath := exePath + ".backup"

	// Remove old backup if exists
	_ = os.Remove(backupPath)

	// Step 1: Move current binary to backup
	if err := os.Rename(exePath, backupPath); err != nil {
		return nil, fmt.Errorf("backup failed: %w", err)
	}

	// Step 2: Move new binary to target location (atomic, same filesystem)
	if err := os.Rename(newBinaryPath, exePath); err != nil {
		// Restore backup on failure
		if restoreErr := os.Rename(backupPath, exePath); restoreErr != nil {
			return nil, fmt.Errorf("replace failed and restore failed: %w (restore error: %v)", err, restoreErr)
		}
		return nil, fmt.Errorf("replace failed (restored backup): %w", err)
	}

	// Success - backup file is kept for rollback capability
	// It will be cleaned up on next successful update
	return &UpdateExecutionResult{
		Message:     "Update completed. Please restart the service.",
		NeedRestart: true,
	}, nil
}

// Rollback restores the previous version
func (s *UpdateService) Rollback() error {
	exePath, err := resolveUpdateExecutablePath()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	backupFile := exePath + ".backup"
	if _, err := os.Stat(backupFile); os.IsNotExist(err) {
		return fmt.Errorf("no backup found")
	}

	// Replace current with backup
	if err := os.Rename(backupFile, exePath); err != nil {
		return fmt.Errorf("rollback failed: %w", err)
	}

	return nil
}

func (s *UpdateService) fetchLatestRelease(ctx context.Context) (*UpdateInfo, error) {
	release, err := s.githubClient.FetchLatestRelease(ctx, githubRepo)
	if err != nil {
		return nil, err
	}

	latestVersion := strings.TrimPrefix(release.TagName, "v")

	assets := make([]Asset, len(release.Assets))
	for i, a := range release.Assets {
		assets[i] = Asset{
			Name:        a.Name,
			DownloadURL: a.BrowserDownloadURL,
			Size:        a.Size,
		}
	}

	return &UpdateInfo{
		CurrentVersion: s.currentVersion,
		LatestVersion:  latestVersion,
		HasUpdate:      compareVersions(s.currentVersion, latestVersion) < 0,
		ReleaseInfo: &ReleaseInfo{
			Name:        release.Name,
			Body:        release.Body,
			PublishedAt: release.PublishedAt,
			HTMLURL:     release.HTMLURL,
			Assets:      assets,
		},
		Cached:    false,
		BuildType: s.buildType,
	}, nil
}

func (s *UpdateService) downloadFile(ctx context.Context, downloadURL, dest string) error {
	return s.githubClient.DownloadFile(ctx, downloadURL, dest, maxDownloadSize)
}

func (s *UpdateService) getArchiveName() string {
	osName := runtime.GOOS
	arch := runtime.GOARCH
	return fmt.Sprintf("%s_%s", osName, arch)
}

// validateDownloadURL checks if the URL is from an allowed domain
// SECURITY: This prevents SSRF and ensures downloads only come from trusted GitHub domains
func validateDownloadURL(rawURL string) error {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	// Must be HTTPS
	if parsedURL.Scheme != "https" {
		return fmt.Errorf("only HTTPS URLs are allowed")
	}

	// Check against allowed hosts
	host := parsedURL.Host
	// GitHub release URLs can be from github.com or objects.githubusercontent.com
	if host != allowedDownloadHost &&
		!strings.HasSuffix(host, "."+allowedDownloadHost) &&
		host != allowedAssetHost &&
		!strings.HasSuffix(host, "."+allowedAssetHost) {
		return fmt.Errorf("download from untrusted host: %s", host)
	}

	return nil
}

func (s *UpdateService) verifyChecksum(ctx context.Context, filePath, checksumURL string) error {
	// Download checksums file
	checksumData, err := s.githubClient.FetchChecksumFile(ctx, checksumURL)
	if err != nil {
		return fmt.Errorf("failed to download checksums: %w", err)
	}

	// Calculate file hash
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return err
	}
	actualHash := hex.EncodeToString(h.Sum(nil))

	// Find expected hash in checksums file
	fileName := filepath.Base(filePath)
	scanner := bufio.NewScanner(strings.NewReader(string(checksumData)))
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) == 2 && parts[1] == fileName {
			if parts[0] == actualHash {
				return nil
			}
			return fmt.Errorf("checksum mismatch: expected %s, got %s", parts[0], actualHash)
		}
	}

	return fmt.Errorf("checksum not found for %s", fileName)
}

func (s *UpdateService) extractBinary(archivePath, destPath string) error {
	f, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	var reader io.Reader = f

	// Handle gzip compression
	if strings.HasSuffix(archivePath, ".gz") || strings.HasSuffix(archivePath, ".tar.gz") || strings.HasSuffix(archivePath, ".tgz") {
		gzr, err := gzip.NewReader(f)
		if err != nil {
			return err
		}
		defer func() { _ = gzr.Close() }()
		reader = gzr
	}

	// Handle tar archive
	if strings.Contains(archivePath, ".tar") {
		tr := tar.NewReader(reader)
		for {
			hdr, err := tr.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				return err
			}

			// SECURITY: Prevent Zip Slip / Path Traversal attack
			// Only allow files with safe base names, no directory traversal
			baseName := filepath.Base(hdr.Name)

			// Check for path traversal attempts
			if strings.Contains(hdr.Name, "..") {
				return fmt.Errorf("path traversal attempt detected: %s", hdr.Name)
			}

			// Validate the entry is a regular file
			if hdr.Typeflag != tar.TypeReg {
				continue // Skip directories and special files
			}

			// Only extract the specific binary we need
			if baseName == "sub2api" || baseName == "sub2api.exe" {
				// Additional security: limit file size (max 500MB)
				const maxBinarySize = 500 * 1024 * 1024
				if hdr.Size > maxBinarySize {
					return fmt.Errorf("binary too large: %d bytes (max %d)", hdr.Size, maxBinarySize)
				}

				out, err := os.Create(destPath)
				if err != nil {
					return err
				}

				// Use LimitReader to prevent decompression bombs
				limited := io.LimitReader(tr, maxBinarySize)
				if _, err := io.Copy(out, limited); err != nil {
					_ = out.Close()
					return err
				}
				if err := out.Close(); err != nil {
					return err
				}
				return nil
			}
		}
		return fmt.Errorf("binary not found in archive")
	}

	// Direct copy for non-tar files (with size limit)
	const maxBinarySize = 500 * 1024 * 1024
	out, err := os.Create(destPath)
	if err != nil {
		return err
	}

	limited := io.LimitReader(reader, maxBinarySize)
	if _, err := io.Copy(out, limited); err != nil {
		_ = out.Close()
		return err
	}
	return out.Close()
}

func (s *UpdateService) getFromCache(ctx context.Context) (*UpdateInfo, error) {
	data, err := s.cache.GetUpdateInfo(ctx)
	if err != nil {
		return nil, err
	}

	var cached struct {
		Latest      string       `json:"latest"`
		ReleaseInfo *ReleaseInfo `json:"release_info"`
		Timestamp   int64        `json:"timestamp"`
	}
	if err := json.Unmarshal([]byte(data), &cached); err != nil {
		return nil, err
	}

	if time.Now().Unix()-cached.Timestamp > updateCacheTTL {
		return nil, fmt.Errorf("cache expired")
	}

	return &UpdateInfo{
		CurrentVersion: s.currentVersion,
		LatestVersion:  cached.Latest,
		HasUpdate:      compareVersions(s.currentVersion, cached.Latest) < 0,
		ReleaseInfo:    cached.ReleaseInfo,
		Cached:         true,
		BuildType:      s.buildType,
	}, nil
}

func (s *UpdateService) saveToCache(ctx context.Context, info *UpdateInfo) {
	cacheData := struct {
		Latest      string       `json:"latest"`
		ReleaseInfo *ReleaseInfo `json:"release_info"`
		Timestamp   int64        `json:"timestamp"`
	}{
		Latest:      info.LatestVersion,
		ReleaseInfo: info.ReleaseInfo,
		Timestamp:   time.Now().Unix(),
	}

	data, _ := json.Marshal(cacheData)
	_ = s.cache.SetUpdateInfo(ctx, string(data), time.Duration(updateCacheTTL)*time.Second)
}

func (s *UpdateService) applyAutoUpdateSupport(info *UpdateInfo) {
	if info == nil {
		return
	}

	canAutoUpdate, updateHint, err := s.detectAutoUpdateSupport(context.Background())
	if err != nil {
		info.CanAutoUpdate = false
		info.Warning = appendUpdateWarning(info.Warning, "Failed to inspect online update support: "+err.Error())
		return
	}

	info.CanAutoUpdate = canAutoUpdate
	info.UpdateHint = updateHint
}

func (s *UpdateService) applyExternalUpdaterRepoStatus(ctx context.Context, info *UpdateInfo) {
	if info == nil || strings.TrimSpace(s.updateConfig.ExternalUpdaterCommand) == "" {
		return
	}

	helperSpec, supported, _, err := s.resolveExternalUpdater(ctx)
	if err != nil {
		info.Warning = appendUpdateWarning(info.Warning, "Failed to inspect updater repository status: "+err.Error())
		return
	}
	if !supported {
		return
	}

	repoStatus, err := inspectExternalUpdaterRepoStatus(ctx, helperSpec.WorkDir)
	if err != nil {
		info.Warning = appendUpdateWarning(info.Warning, "Failed to inspect updater repository status: "+err.Error())
		return
	}

	if repoStatus.BehindCount <= 0 {
		info.HasUpdate = false
		if strings.TrimSpace(info.CurrentVersion) != "" {
			info.LatestVersion = info.CurrentVersion
			return
		}
		currentLabel := formatExternalUpdaterVersionLabel(repoStatus.CurrentRef, repoStatus.CurrentCommit)
		info.CurrentVersion = currentLabel
		info.LatestVersion = currentLabel
		return
	}

	info.HasUpdate = true
	if compareVersions(info.CurrentVersion, info.LatestVersion) >= 0 {
		info.LatestVersion = formatExternalUpdaterVersionLabel(repoStatus.UpstreamRef, repoStatus.UpstreamCommit)
	}
}

func (s *UpdateService) ensureAutoUpdateSupported(ctx context.Context) error {
	canAutoUpdate, updateHint, err := s.detectAutoUpdateSupport(ctx)
	if err != nil {
		return infraerrors.InternalServer(
			autoUpdateCheckFailedReason,
			"无法检查当前部署是否支持在线更新，请查看服务日志 (failed to inspect whether online update is supported in this deployment)",
		).WithCause(err)
	}
	if canAutoUpdate {
		return nil
	}
	return infraerrors.Conflict(autoUpdateUnsupportedReason, updateHint)
}

func (s *UpdateService) detectAutoUpdateSupport(ctx context.Context) (bool, string, error) {
	if _, supported, supportHint, err := s.resolveExternalUpdater(ctx); err != nil {
		return false, "", err
	} else if supported {
		return true, "", nil
	} else if supportHint != "" && strings.TrimSpace(s.updateConfig.ExternalUpdaterCommand) != "" {
		return false, supportHint, nil
	}

	if s.buildType != "release" {
		return false, sourceBuildUpdateHint, nil
	}

	exePath, err := resolveUpdateExecutablePath()
	if err != nil {
		return false, "", fmt.Errorf("failed to resolve executable path: %w", err)
	}

	return probeAutoUpdateWorkspace(exePath)
}

func probeAutoUpdateWorkspace(exePath string) (bool, string, error) {
	if strings.TrimSpace(exePath) == "" {
		return false, "", fmt.Errorf("executable path is empty")
	}

	tempDir, err := createUpdateTempDir(filepath.Dir(exePath), ".sub2api-update-*")
	if err != nil {
		if errors.Is(err, fs.ErrPermission) {
			return false, manualUpdateRequiredHint, nil
		}
		return false, "", fmt.Errorf("failed to create update workspace: %w", err)
	}

	_ = removeUpdatePathAll(tempDir)
	return true, "", nil
}

func (s *UpdateService) resolveExternalUpdater(ctx context.Context) (updateExternalHelperSpec, bool, string, error) {
	command := strings.TrimSpace(s.updateConfig.ExternalUpdaterCommand)
	if command == "" {
		return updateExternalHelperSpec{}, false, "", nil
	}

	commandPath, err := filepath.Abs(command)
	if err != nil {
		return updateExternalHelperSpec{}, false, formatExternalUpdaterHint(fmt.Errorf("invalid updater command path: %w", err)), nil
	}

	info, err := os.Stat(commandPath)
	if err != nil {
		return updateExternalHelperSpec{}, false, formatExternalUpdaterHint(fmt.Errorf("updater command not found: %w", err)), nil
	}
	if info.IsDir() || info.Mode().Perm()&0o111 == 0 {
		return updateExternalHelperSpec{}, false, formatExternalUpdaterHint(fmt.Errorf("updater command is not executable: %s", commandPath)), nil
	}

	workDir := strings.TrimSpace(s.updateConfig.ExternalUpdaterWorkingDirectory)
	if workDir == "" {
		workDir = filepath.Dir(commandPath)
	}
	if !filepath.IsAbs(workDir) {
		return updateExternalHelperSpec{}, false, formatExternalUpdaterHint(fmt.Errorf("updater working directory must be an absolute path: %s", workDir)), nil
	}
	if dirInfo, err := os.Stat(workDir); err != nil || !dirInfo.IsDir() {
		if err == nil {
			err = fmt.Errorf("%s is not a directory", workDir)
		}
		return updateExternalHelperSpec{}, false, formatExternalUpdaterHint(fmt.Errorf("updater working directory unavailable: %w", err)), nil
	}

	if err := runUpdateDockerAccessCheck(ctx); err != nil {
		return updateExternalHelperSpec{}, false, formatExternalUpdaterHint(err), nil
	}

	helperImage, err := resolveUpdateHelperImage(ctx)
	if err != nil {
		return updateExternalHelperSpec{}, false, formatExternalUpdaterHint(err), nil
	}

	timeout := time.Duration(s.updateConfig.ExternalUpdaterHelperTimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 30 * time.Minute
	}

	return updateExternalHelperSpec{
		HelperImage: helperImage,
		CommandPath: commandPath,
		WorkDir:     workDir,
		Timeout:     timeout,
	}, true, "", nil
}

func defaultRunUpdateDockerAccessCheck(ctx context.Context) error {
	checkCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(checkCtx, "docker", "info", "--format", "{{.ServerVersion}}")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("docker access unavailable: %w: %s", err, trimUpdateHelperOutput(output))
	}
	return nil
}

func defaultResolveUpdateHelperImage(ctx context.Context) (string, error) {
	checkCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	containerID, err := os.Hostname()
	if err != nil {
		return "", fmt.Errorf("resolve current container hostname: %w", err)
	}
	containerID = strings.TrimSpace(containerID)
	if containerID == "" {
		return "", fmt.Errorf("current container hostname is empty")
	}

	cmd := exec.CommandContext(checkCtx, "docker", "inspect", containerID, "--format", "{{.Config.Image}}")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("inspect current container image: %w: %s", err, trimUpdateHelperOutput(output))
	}

	image := strings.TrimSpace(string(output))
	if image == "" {
		return "", fmt.Errorf("current container image is empty")
	}
	return image, nil
}

func defaultLaunchDetachedExternalUpdater(ctx context.Context, spec updateExternalHelperSpec) error {
	timeout := int(spec.Timeout / time.Second)
	if timeout <= 0 {
		timeout = int((30 * time.Minute) / time.Second)
	}

	name := fmt.Sprintf("sub2api-update-%d", time.Now().UnixNano())
	cmd := exec.CommandContext(
		ctx,
		"docker",
		"run",
		"-d",
		"--rm",
		"--name", name,
		"-v", "/var/run/docker.sock:/var/run/docker.sock",
		"-v", fmt.Sprintf("%s:%s", spec.WorkDir, spec.WorkDir),
		"-w", spec.WorkDir,
		"-e", fmt.Sprintf("UPDATE_HELPER_TIMEOUT_SECONDS=%d", timeout),
		spec.HelperImage,
		spec.CommandPath,
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("start update helper container: %w: %s", err, trimUpdateHelperOutput(output))
	}
	return nil
}

func defaultInspectExternalUpdaterRepoStatus(ctx context.Context, workDir string) (*externalUpdaterRepoStatus, error) {
	workDir = strings.TrimSpace(workDir)
	if workDir == "" {
		return nil, fmt.Errorf("external updater working directory is empty")
	}

	fetchCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	if _, err := runGitInWorkDir(fetchCtx, workDir, "fetch", "--quiet", "--prune"); err != nil {
		return nil, err
	}

	currentRef, err := runGitInWorkDir(fetchCtx, workDir, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return nil, err
	}
	currentCommit, err := runGitInWorkDir(fetchCtx, workDir, "rev-parse", "HEAD")
	if err != nil {
		return nil, err
	}
	upstreamRef, err := runGitInWorkDir(fetchCtx, workDir, "rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{u}")
	if err != nil {
		return nil, err
	}
	upstreamCommit, err := runGitInWorkDir(fetchCtx, workDir, "rev-parse", "@{u}")
	if err != nil {
		return nil, err
	}
	behindCountRaw, err := runGitInWorkDir(fetchCtx, workDir, "rev-list", "--count", "HEAD..@{u}")
	if err != nil {
		return nil, err
	}
	behindCount, err := strconv.Atoi(strings.TrimSpace(behindCountRaw))
	if err != nil {
		return nil, fmt.Errorf("parse updater repository behind count: %w", err)
	}

	return &externalUpdaterRepoStatus{
		CurrentRef:     strings.TrimSpace(currentRef),
		CurrentCommit:  strings.TrimSpace(currentCommit),
		UpstreamRef:    strings.TrimSpace(upstreamRef),
		UpstreamCommit: strings.TrimSpace(upstreamCommit),
		BehindCount:    behindCount,
	}, nil
}

func runGitInWorkDir(ctx context.Context, workDir string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", append([]string{"-C", workDir}, args...)...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git %s failed: %w: %s", strings.Join(args, " "), err, trimUpdateHelperOutput(output))
	}
	return strings.TrimSpace(string(output)), nil
}

func formatExternalUpdaterHint(err error) string {
	return "已配置在线更新器，但当前容器无法触发宿主机更新，请检查 docker.sock 挂载、权限和仓库挂载: " +
		strings.TrimSpace(err.Error()) +
		" (configured online updater is unavailable; verify docker.sock permissions and repository bind mounts)"
}

func trimUpdateHelperOutput(output []byte) string {
	msg := strings.TrimSpace(string(output))
	if msg == "" {
		return "no command output"
	}
	const maxLen = 240
	if len(msg) <= maxLen {
		return msg
	}
	return msg[:maxLen] + "..."
}

func appendUpdateWarning(current, next string) string {
	next = strings.TrimSpace(next)
	if next == "" {
		return current
	}
	if strings.TrimSpace(current) == "" {
		return next
	}
	return current + "; " + next
}

func formatExternalUpdaterVersionLabel(refName, commit string) string {
	refName = strings.TrimSpace(refName)
	shortCommit := shortenGitCommit(commit)
	if refName == "" {
		return shortCommit
	}
	if shortCommit == "" {
		return refName
	}
	return refName + "@" + shortCommit
}

func shortenGitCommit(commit string) string {
	commit = strings.TrimSpace(commit)
	if len(commit) <= 8 {
		return commit
	}
	return commit[:8]
}

// compareVersions compares two semantic versions
func compareVersions(current, latest string) int {
	currentParts := parseVersion(current)
	latestParts := parseVersion(latest)

	for i := 0; i < 3; i++ {
		if currentParts[i] < latestParts[i] {
			return -1
		}
		if currentParts[i] > latestParts[i] {
			return 1
		}
	}
	return 0
}

func parseVersion(v string) [3]int {
	v = strings.TrimPrefix(v, "v")
	parts := strings.Split(v, ".")
	result := [3]int{0, 0, 0}
	for i := 0; i < len(parts) && i < 3; i++ {
		if parsed, err := strconv.Atoi(parts[i]); err == nil {
			result[i] = parsed
		}
	}
	return result
}
