//go:build unit

package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestAccountIsSchedulable_IgnoresOpenAIOAuthTokenDerivedExpiry(t *testing.T) {
	past := time.Now().Add(-1 * time.Hour).UTC().Truncate(time.Second)
	account := &Account{
		Platform:           PlatformOpenAI,
		Type:               AccountTypeOAuth,
		Status:             StatusActive,
		Schedulable:        true,
		AutoPauseOnExpired: true,
		ExpiresAt:          &past,
		Credentials: map[string]any{
			"expires_at": past.Format(time.RFC3339),
		},
	}

	require.True(t, account.HasTokenDerivedAccountExpiry())
	require.False(t, account.IsAccountExpiryExpired(time.Now()))
	require.True(t, account.IsSchedulable())
}

func TestAccountIsSchedulable_BlocksManualExpiredAccountExpiry(t *testing.T) {
	past := time.Now().Add(-1 * time.Hour).UTC().Truncate(time.Second)
	account := &Account{
		Platform:           PlatformOpenAI,
		Type:               AccountTypeOAuth,
		Status:             StatusActive,
		Schedulable:        true,
		AutoPauseOnExpired: true,
		ExpiresAt:          &past,
		Credentials: map[string]any{
			"expires_at": past.Add(10 * time.Minute).Format(time.RFC3339),
		},
	}

	require.False(t, account.HasTokenDerivedAccountExpiry())
	require.True(t, account.IsAccountExpiryExpired(time.Now()))
	require.False(t, account.IsSchedulable())
}

func TestAccountIsSchedulable_OnlyOpenAIOAuthUsesTokenDerivedExpiry(t *testing.T) {
	past := time.Now().Add(-1 * time.Hour).UTC().Truncate(time.Second)
	account := &Account{
		Platform:           PlatformAnthropic,
		Type:               AccountTypeOAuth,
		Status:             StatusActive,
		Schedulable:        true,
		AutoPauseOnExpired: true,
		ExpiresAt:          &past,
		Credentials: map[string]any{
			"expires_at": past.Format(time.RFC3339),
		},
	}

	require.False(t, account.HasTokenDerivedAccountExpiry())
	require.True(t, account.IsAccountExpiryExpired(time.Now()))
	require.False(t, account.IsSchedulable())
}
