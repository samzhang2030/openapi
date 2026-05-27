package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAccount_IsAnthropicAPIKeyPassthroughEnabled(t *testing.T) {
	t.Run("Anthropic API Key 开启", func(t *testing.T) {
		account := &Account{
			Platform: PlatformAnthropic,
			Type:     AccountTypeAPIKey,
			Extra: map[string]any{
				"anthropic_passthrough": true,
			},
		}
		require.True(t, account.IsAnthropicAPIKeyPassthroughEnabled())
	})

	t.Run("Anthropic API Key 关闭", func(t *testing.T) {
		account := &Account{
			Platform: PlatformAnthropic,
			Type:     AccountTypeAPIKey,
			Extra: map[string]any{
				"anthropic_passthrough": false,
			},
		}
		require.False(t, account.IsAnthropicAPIKeyPassthroughEnabled())
	})

	t.Run("字段类型非法默认关闭", func(t *testing.T) {
		account := &Account{
			Platform: PlatformAnthropic,
			Type:     AccountTypeAPIKey,
			Extra: map[string]any{
				"anthropic_passthrough": "true",
			},
		}
		require.False(t, account.IsAnthropicAPIKeyPassthroughEnabled())
	})

	t.Run("非 Anthropic API Key 账号始终关闭", func(t *testing.T) {
		oauth := &Account{
			Platform: PlatformAnthropic,
			Type:     AccountTypeOAuth,
			Extra: map[string]any{
				"anthropic_passthrough": true,
			},
		}
		require.False(t, oauth.IsAnthropicAPIKeyPassthroughEnabled())

		openai := &Account{
			Platform: PlatformOpenAI,
			Type:     AccountTypeAPIKey,
			Extra: map[string]any{
				"anthropic_passthrough": true,
			},
		}
		require.False(t, openai.IsAnthropicAPIKeyPassthroughEnabled())
	})

	t.Run("Right Code 上游默认开启透传", func(t *testing.T) {
		account := &Account{
			Platform: PlatformAnthropic,
			Type:     AccountTypeAPIKey,
			Credentials: map[string]any{
				"base_url": "https://www.right.codes",
			},
		}
		require.True(t, account.IsAnthropicAPIKeyPassthroughEnabled())
	})

	t.Run("Right Code 显式关闭优先", func(t *testing.T) {
		account := &Account{
			Platform: PlatformAnthropic,
			Type:     AccountTypeAPIKey,
			Credentials: map[string]any{
				"base_url": "https://api.right.codes",
			},
			Extra: map[string]any{
				"anthropic_passthrough": false,
			},
		}
		require.False(t, account.IsAnthropicAPIKeyPassthroughEnabled())
	})
}
