//go:build unit

package repository

import (
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestBuildSchedulerMetadataAccount_KeepsOpenAIWSFlags(t *testing.T) {
	account := service.Account{
		ID:       42,
		Platform: service.PlatformOpenAI,
		Type:     service.AccountTypeOAuth,
		Extra: map[string]any{
			"openai_oauth_responses_websockets_v2_enabled": true,
			"openai_oauth_responses_websockets_v2_mode":    service.OpenAIWSIngressModePassthrough,
			"openai_ws_force_http":                         true,
			"openai_responses_mode":                        "force_chat_completions",
			"openai_responses_supported":                   false,
			"mixed_scheduling":                             true,
			"unused_large_field":                           "drop-me",
		},
	}

	got := buildSchedulerMetadataAccount(account)

	require.Equal(t, true, got.Extra["openai_oauth_responses_websockets_v2_enabled"])
	require.Equal(t, service.OpenAIWSIngressModePassthrough, got.Extra["openai_oauth_responses_websockets_v2_mode"])
	require.Equal(t, true, got.Extra["openai_ws_force_http"])
	require.Equal(t, "force_chat_completions", got.Extra["openai_responses_mode"])
	require.Equal(t, false, got.Extra["openai_responses_supported"])
	require.Equal(t, true, got.Extra["mixed_scheduling"])
	require.Nil(t, got.Extra["unused_large_field"])
}

func TestBuildSchedulerMetadataAccount_KeepsQuotaFields(t *testing.T) {
	now := time.Now().UTC()
	account := service.Account{
		ID:       42,
		Platform: service.PlatformOpenAI,
		Type:     service.AccountTypeAPIKey,
		Status:   service.StatusActive,
		Extra: map[string]any{
			"quota_limit":             200.0,
			"quota_used":              10.0,
			"quota_daily_limit":       20.0,
			"quota_daily_used":        20.0,
			"quota_daily_start":       now.Format(time.RFC3339),
			"quota_daily_reset_mode":  "fixed",
			"quota_daily_reset_hour":  9.0,
			"quota_weekly_limit":      100.0,
			"quota_weekly_used":       40.0,
			"quota_weekly_start":      now.Format(time.RFC3339),
			"quota_weekly_reset_mode": "fixed",
			"quota_weekly_reset_day":  1.0,
			"quota_weekly_reset_hour": 9.0,
			"quota_reset_timezone":    "UTC",
			"unused_large_field":      "drop-me",
		},
	}

	got := buildSchedulerMetadataAccount(account)

	require.Equal(t, 200.0, got.Extra["quota_limit"])
	require.Equal(t, 20.0, got.Extra["quota_daily_limit"])
	require.Equal(t, 20.0, got.Extra["quota_daily_used"])
	require.Equal(t, now.Format(time.RFC3339), got.Extra["quota_daily_start"])
	require.Equal(t, "fixed", got.Extra["quota_daily_reset_mode"])
	require.Equal(t, 9.0, got.Extra["quota_daily_reset_hour"])
	require.Equal(t, 100.0, got.Extra["quota_weekly_limit"])
	require.Equal(t, 40.0, got.Extra["quota_weekly_used"])
	require.Equal(t, "UTC", got.Extra["quota_reset_timezone"])
	require.Nil(t, got.Extra["unused_large_field"])
	require.True(t, got.IsQuotaExceeded(), "scheduler metadata must preserve quota fields used by IsSchedulable")
}

func TestBuildSchedulerMetadataAccount_KeepsSlimGroupMembership(t *testing.T) {
	account := service.Account{
		ID:       42,
		Platform: service.PlatformAnthropic,
		GroupIDs: []int64{7, 9, 7, 0},
		AccountGroups: []service.AccountGroup{
			{
				AccountID: 42,
				GroupID:   7,
				Priority:  2,
				Account:   &service.Account{ID: 42, Name: "drop-from-metadata"},
				Group:     &service.Group{ID: 7, Name: "drop-from-metadata"},
			},
			{
				AccountID: 42,
				GroupID:   11,
				Priority:  3,
				Group:     &service.Group{ID: 11, Name: "drop-from-metadata"},
			},
			{
				AccountID: 42,
				GroupID:   0,
				Priority:  4,
			},
		},
	}

	got := buildSchedulerMetadataAccount(account)

	require.Equal(t, []int64{7, 9, 11}, got.GroupIDs)
	require.Len(t, got.AccountGroups, 2)
	require.Equal(t, int64(42), got.AccountGroups[0].AccountID)
	require.Equal(t, int64(7), got.AccountGroups[0].GroupID)
	require.Equal(t, 2, got.AccountGroups[0].Priority)
	require.Nil(t, got.AccountGroups[0].Account)
	require.Nil(t, got.AccountGroups[0].Group)
	require.Equal(t, int64(11), got.AccountGroups[1].GroupID)
	require.Nil(t, got.Groups)
}
