package service

import (
	"bytes"
	"context"
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

type tlsRecordingUpstream struct {
	doCalls        int
	doWithTLSCalls int
	lastReq        *http.Request
	lastBody       []byte
	lastProfile    *tlsfingerprint.Profile
	resp           *http.Response
	err            error
}

func (u *tlsRecordingUpstream) Do(_ *http.Request, _ string, _ int64, _ int) (*http.Response, error) {
	u.doCalls++
	return nil, fmt.Errorf("unexpected Do call")
}

func (u *tlsRecordingUpstream) DoWithTLS(req *http.Request, _ string, _ int64, _ int, profile *tlsfingerprint.Profile) (*http.Response, error) {
	u.doWithTLSCalls++
	u.lastReq = req
	u.lastProfile = profile
	if req != nil && req.Body != nil {
		body, _ := io.ReadAll(req.Body)
		u.lastBody = body
		_ = req.Body.Close()
		req.Body = io.NopCloser(bytes.NewReader(body))
	}
	if u.err != nil {
		return nil, u.err
	}
	return u.resp, nil
}

func TestOpenAIGatewayService_Forward_OpenAIOAuthUsesTLSFingerprint(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", bytes.NewReader(nil))

	upstream := &tlsRecordingUpstream{
		resp: &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"text/event-stream"}, "x-request-id": []string{"rid"}},
			Body: io.NopCloser(strings.NewReader(strings.Join([]string{
				`data: {"type":"response.completed","response":{"usage":{"input_tokens":1,"output_tokens":1,"input_tokens_details":{"cached_tokens":0}}}}`,
				"",
				"data: [DONE]",
				"",
			}, "\n"))),
		},
	}

	svc := &OpenAIGatewayService{
		cfg:          &config.Config{},
		httpUpstream: upstream,
	}

	account := &Account{
		ID:          1,
		Name:        "openai-oauth",
		Platform:    PlatformOpenAI,
		Type:        AccountTypeOAuth,
		Concurrency: 1,
		Credentials: map[string]any{
			"access_token":       "oauth-token",
			"chatgpt_account_id": "chatgpt-account",
		},
		Extra: map[string]any{
			"enable_tls_fingerprint": true,
		},
	}

	result, err := svc.Forward(context.Background(), c, account, []byte(`{"model":"gpt-4.1-mini","stream":false,"input":"Reply with exactly: ok"}`))
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Zero(t, upstream.doCalls)
	require.Equal(t, 1, upstream.doWithTLSCalls)
	require.NotNil(t, upstream.lastProfile)
	require.Equal(t, "Built-in Default (Node.js 24.x)", upstream.lastProfile.Name)
	require.Equal(t, "Bearer oauth-token", upstream.lastReq.Header.Get("Authorization"))
	require.Equal(t, "gpt-5.4-mini", gjson.GetBytes(upstream.lastBody, "model").String())
}

func TestOpenAIGatewayService_Forward_OpenAIOAuthPassthroughUsesTLSFingerprint(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", bytes.NewReader(nil))
	c.Request.Header.Set("User-Agent", "codex_cli_rs/0.1.0")

	upstream := &tlsRecordingUpstream{
		resp: &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}, "x-request-id": []string{"rid"}},
			Body:       io.NopCloser(strings.NewReader(`{"output":[],"usage":{"input_tokens":1,"output_tokens":1,"input_tokens_details":{"cached_tokens":0}}}`)),
		},
	}

	svc := &OpenAIGatewayService{
		cfg:          &config.Config{},
		httpUpstream: upstream,
	}

	account := &Account{
		ID:          2,
		Name:        "openai-oauth-passthrough",
		Platform:    PlatformOpenAI,
		Type:        AccountTypeOAuth,
		Concurrency: 1,
		Credentials: map[string]any{
			"access_token":       "oauth-token",
			"chatgpt_account_id": "chatgpt-account",
		},
		Extra: map[string]any{
			"enable_tls_fingerprint": true,
			"openai_passthrough":     true,
		},
	}

	_, err := svc.Forward(context.Background(), c, account, []byte(`{"model":"gpt-5.2","stream":true,"input":[{"type":"text","text":"hi"}]}`))
	require.ErrorContains(t, err, "missing terminal event")
	require.Zero(t, upstream.doCalls)
	require.Equal(t, 1, upstream.doWithTLSCalls)
	require.NotNil(t, upstream.lastProfile)
	require.Equal(t, "Built-in Default (Node.js 24.x)", upstream.lastProfile.Name)
	require.Equal(t, "Bearer oauth-token", upstream.lastReq.Header.Get("Authorization"))
}
