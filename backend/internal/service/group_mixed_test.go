package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolveMixedModelPlatform(t *testing.T) {
	tests := []struct {
		name     string
		model    string
		expected string
	}{
		{name: "deepseek", model: "deepseek-chat", expected: PlatformDeepSeek},
		{name: "gemini", model: "gemini-2.5-pro", expected: PlatformGemini},
		{name: "claude", model: "claude-sonnet-4-6", expected: PlatformAnthropic},
		{name: "gpt", model: "gpt-5.4", expected: PlatformOpenAI},
		{name: "codex", model: "codex-mini-latest", expected: PlatformOpenAI},
		{name: "unknown defaults openai", model: "future-model", expected: PlatformOpenAI},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, ResolveMixedModelPlatform(tt.model))
		})
	}
}
