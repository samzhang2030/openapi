//go:build unit

package service

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsOpenAIPrivacyCloudflareChallenge(t *testing.T) {
	t.Parallel()

	headers := http.Header{}
	headers.Set("Content-Type", "text/html")
	headers.Set("Cf-Ray", "test-ray")

	body := `<!DOCTYPE html><title>Just a moment...</title><script>window._cf_chl_opt={};</script><div>Enable JavaScript and cookies to continue</div>`

	require.True(t, isOpenAIPrivacyCloudflareChallenge(http.StatusForbidden, headers, body))
}

func TestIsOpenAIPrivacyCloudflareChallenge_ServiceUnavailable(t *testing.T) {
	t.Parallel()

	headers := http.Header{}
	headers.Set("Content-Type", "text/html")

	body := `<!DOCTYPE html><html><body><script>window._cf_chl_opt={};</script></body></html>`

	require.True(t, isOpenAIPrivacyCloudflareChallenge(http.StatusServiceUnavailable, headers, body))
}

func TestIsOpenAIPrivacyCloudflareChallenge_DoesNotMatchRegularErrors(t *testing.T) {
	t.Parallel()

	headers := http.Header{}
	headers.Set("Content-Type", "application/json")

	body := `{"error":{"message":"token expired"}}`

	require.False(t, isOpenAIPrivacyCloudflareChallenge(http.StatusForbidden, headers, body))
	require.False(t, isOpenAIPrivacyCloudflareChallenge(http.StatusInternalServerError, headers, body))
}
