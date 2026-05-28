package handler

import (
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/pkg/openai"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestDefaultModelsForPlatform_DeepSeek(t *testing.T) {
	models, ok := defaultModelsForPlatform(service.PlatformDeepSeek).([]openai.Model)
	require.True(t, ok)

	ids := make(map[string]struct{}, len(models))
	for _, model := range models {
		ids[model.ID] = struct{}{}
	}

	require.Contains(t, ids, "deepseek-v4-pro")
	require.Contains(t, ids, "deepseek-v4-flash")
	require.NotContains(t, ids, "claude-sonnet-4-6")
}

func TestDefaultModelsForPlatform_MixedContainsAllProviderFamilies(t *testing.T) {
	models, ok := defaultModelsForPlatform(service.PlatformMixed).([]gin.H)
	require.True(t, ok)

	ids := make(map[string]struct{}, len(models))
	for _, model := range models {
		id, _ := model["id"].(string)
		if id != "" {
			ids[id] = struct{}{}
		}
	}

	require.Contains(t, ids, "gpt-5.5")
	require.Contains(t, ids, "gpt-5.4")
	require.Contains(t, ids, "claude-sonnet-4-6")
	require.Contains(t, ids, "gemini-2.0-flash")
	require.Contains(t, ids, "deepseek-v4-pro")
}

func TestModelsForListResponse_MixedMergesMappedModelsWithDefaults(t *testing.T) {
	models, ok := modelsForListResponse(service.PlatformMixed, []string{"gpt-5.5", "deepseek-chat"}).([]gin.H)
	require.True(t, ok)

	ids := make(map[string]struct{}, len(models))
	for _, model := range models {
		id, _ := model["id"].(string)
		if id != "" {
			ids[id] = struct{}{}
		}
	}

	require.Contains(t, ids, "gpt-5.5")
	require.Contains(t, ids, "claude-sonnet-4-6")
	require.Contains(t, ids, "gemini-2.5-pro")
	require.Contains(t, ids, "deepseek-v4-pro")
	require.NotContains(t, ids, "deepseek-chat")
}

func TestModelsForListResponse_MixedUsesOpenAICompatibleModelShape(t *testing.T) {
	models, ok := modelsForListResponse(service.PlatformMixed, []string{"custom-model"}).([]gin.H)
	require.True(t, ok)

	for _, model := range models {
		id, _ := model["id"].(string)
		require.NotEmpty(t, id)
		require.Equal(t, "model", model["object"])
		require.NotEmpty(t, model["owned_by"], id)
		require.NotZero(t, model["created"], id)
	}
}
