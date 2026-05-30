//go:build unit

package service

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestAccountTestService_ClaudeAPIKeyTestSetsSessionHeader(t *testing.T) {
	c, _ := newTestContext()
	resp := newJSONResponse(http.StatusOK, "data: {\"type\":\"message_stop\"}\n\n")
	upstream := &anthropicHTTPUpstreamRecorder{resp: resp}
	svc := &AccountTestService{
		httpUpstream: upstream,
		cfg: &config.Config{
			Security: config.SecurityConfig{
				URLAllowlist: config.URLAllowlistConfig{
					Enabled: false,
				},
			},
		},
	}
	account := &Account{
		ID:       44,
		Platform: PlatformAnthropic,
		Type:     AccountTypeAPIKey,
		Credentials: map[string]any{
			"api_key":  "right-key",
			"base_url": "https://www.right.codes",
		},
	}

	err := svc.testClaudeAccountConnection(c, account, "claude-opus-4-5-20251101")
	require.NoError(t, err)
	require.NotNil(t, upstream.lastReq)

	sessionHeader := upstream.lastReq.Header.Get("X-Claude-Code-Session-Id")
	require.NotEmpty(t, sessionHeader)

	userID := gjson.GetBytes(upstream.lastBody, "metadata.user_id").String()
	parsed := ParseMetadataUserID(userID)
	require.NotNil(t, parsed)
	require.Equal(t, parsed.SessionID, sessionHeader)
	require.Equal(t, "right-key", upstream.lastReq.Header.Get("x-api-key"))

	body, err := io.ReadAll(upstream.lastReq.Body)
	require.NoError(t, err)
	require.True(t, strings.Contains(string(body), "claude-opus-4-5-20251101"))
}
