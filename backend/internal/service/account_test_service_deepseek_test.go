package service

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/tlsfingerprint"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

type deepSeekTestUpstream struct {
	response *http.Response
	request  *http.Request
}

func (u *deepSeekTestUpstream) Do(_ *http.Request, _ string, _ int64, _ int) (*http.Response, error) {
	return nil, fmt.Errorf("unexpected Do call")
}

func (u *deepSeekTestUpstream) DoWithTLS(req *http.Request, _ string, _ int64, _ int, _ *tlsfingerprint.Profile) (*http.Response, error) {
	u.request = req
	return u.response, nil
}

func newDeepSeekTestContext() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/admin/accounts/1/test", nil)
	return c, rec
}

func TestAccountTestService_DeepSeekUsesChatCompletions(t *testing.T) {
	ctx, recorder := newDeepSeekTestContext()
	upstream := &deepSeekTestUpstream{
		response: &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body: io.NopCloser(strings.NewReader(`data: {"choices":[{"delta":{"content":"ok"}}]}
data: [DONE]

`)),
		},
	}
	svc := &AccountTestService{httpUpstream: upstream, cfg: &config.Config{}}
	account := &Account{
		ID:          45,
		Platform:    PlatformDeepSeek,
		Type:        AccountTypeAPIKey,
		Concurrency: 1,
		Credentials: map[string]any{"api_key": "sk-test", "base_url": "https://api.deepseek.com"},
	}

	err := svc.testDeepSeekAccountConnection(ctx, account, "")
	require.NoError(t, err)
	require.NotNil(t, upstream.request)
	require.Equal(t, "https://api.deepseek.com/v1/chat/completions", upstream.request.URL.String())
	require.Equal(t, "Bearer sk-test", upstream.request.Header.Get("Authorization"))

	body, readErr := io.ReadAll(upstream.request.Body)
	require.NoError(t, readErr)
	require.Equal(t, "deepseek-v4-pro", gjson.GetBytes(body, "model").String())
	require.True(t, gjson.GetBytes(body, "stream").Bool())
	require.Contains(t, recorder.Body.String(), `"text":"ok"`)
	require.Contains(t, recorder.Body.String(), `"type":"test_complete"`)
}
