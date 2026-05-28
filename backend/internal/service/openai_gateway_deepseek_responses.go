package service

import (
	"context"

	"github.com/gin-gonic/gin"
)

// ForwardDeepSeekAsResponses bridges OpenAI Responses API clients to DeepSeek's
// OpenAI-compatible Chat Completions upstream.
func (s *OpenAIGatewayService) ForwardDeepSeekAsResponses(
	ctx context.Context,
	c *gin.Context,
	account *Account,
	body []byte,
	_ string,
) (*OpenAIForwardResult, error) {
	return s.forwardResponsesViaRawChatCompletions(ctx, c, account, body)
}
