package routes

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestMixedGatewayRouteHelpersRouteByModelAndRestoreBody(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		body           string
		chatOpenAI     bool
		messageOpenAI  bool
		responseOpenAI bool
	}{
		{name: "gpt", body: `{"model":"gpt-5.4","messages":[]}`, chatOpenAI: true, messageOpenAI: true, responseOpenAI: true},
		{name: "deepseek", body: `{"model":"deepseek-chat","messages":[]}`, chatOpenAI: true, messageOpenAI: true, responseOpenAI: true},
		{name: "claude", body: `{"model":"claude-sonnet-4-6","messages":[]}`, chatOpenAI: false, messageOpenAI: false, responseOpenAI: false},
		{name: "gemini", body: `{"model":"gemini-2.5-pro","messages":[]}`, chatOpenAI: false, messageOpenAI: false, responseOpenAI: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newMixedGatewayRouteTestContext(tt.body)

			require.Equal(t, tt.chatOpenAI, shouldUseOpenAICompatibleChatGateway(c))
			bodyAfterChat, err := io.ReadAll(c.Request.Body)
			require.NoError(t, err)
			require.Equal(t, tt.body, string(bodyAfterChat))

			c.Request.Body = io.NopCloser(strings.NewReader(tt.body))
			require.Equal(t, tt.messageOpenAI, shouldUseOpenAICompatibleMessagesGateway(c))
			bodyAfterMessages, err := io.ReadAll(c.Request.Body)
			require.NoError(t, err)
			require.Equal(t, tt.body, string(bodyAfterMessages))

			c.Request.Body = io.NopCloser(strings.NewReader(tt.body))
			require.Equal(t, tt.responseOpenAI, shouldUseOpenAICompatibleResponsesGateway(c))
			bodyAfterResponses, err := io.ReadAll(c.Request.Body)
			require.NoError(t, err)
			require.Equal(t, tt.body, string(bodyAfterResponses))
		})
	}
}

func newMixedGatewayRouteTestContext(body string) *gin.Context {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(body))
	c.Set(string(middleware.ContextKeyAPIKey), &service.APIKey{
		Group: &service.Group{Platform: service.PlatformMixed},
	})
	return c
}
