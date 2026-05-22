//go:build unit

package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

type accountRepoStubForCreateAccountDefaults struct {
	accountRepoStub
	created   *Account
	bindCalls []accountCreateDefaultsBindCall
}

type accountCreateDefaultsBindCall struct {
	accountID int64
	groupIDs  []int64
}

func (s *accountRepoStubForCreateAccountDefaults) Create(_ context.Context, account *Account) error {
	account.ID = 900
	clone := *account
	if account.Extra != nil {
		clone.Extra = make(map[string]any, len(account.Extra))
		for k, v := range account.Extra {
			clone.Extra[k] = v
		}
	}
	s.created = &clone
	return nil
}

func (s *accountRepoStubForCreateAccountDefaults) BindGroups(_ context.Context, accountID int64, groupIDs []int64) error {
	s.bindCalls = append(s.bindCalls, accountCreateDefaultsBindCall{
		accountID: accountID,
		groupIDs:  append([]int64(nil), groupIDs...),
	})
	return nil
}

type groupRepoStubForCreateAccountDefaults struct {
	groupRepoStub
	activeByPlatform map[string][]Group
	active           []Group
}

func (s *groupRepoStubForCreateAccountDefaults) ListActiveByPlatform(_ context.Context, platform string) ([]Group, error) {
	return append([]Group(nil), s.activeByPlatform[platform]...), nil
}

func (s *groupRepoStubForCreateAccountDefaults) ListActive(context.Context) ([]Group, error) {
	return append([]Group(nil), s.active...), nil
}

func TestAdminServiceCreateAccountOpenAIAPIKeyDefaultsWSAndMixedDefaultGroup(t *testing.T) {
	accountRepo := &accountRepoStubForCreateAccountDefaults{}
	groupRepo := &groupRepoStubForCreateAccountDefaults{
		active: []Group{{ID: 12, Name: openAIAPIKeyDefaultGroupName, Platform: "mixed", Status: StatusActive}},
	}
	svc := &adminServiceImpl{accountRepo: accountRepo, groupRepo: groupRepo}

	_, err := svc.CreateAccount(context.Background(), &CreateAccountInput{
		Name:     "openai-apikey",
		Platform: PlatformOpenAI,
		Type:     AccountTypeAPIKey,
		GroupIDs: []int64{4, 5},
		Extra: map[string]any{
			"openai_apikey_responses_websockets_v2_mode":    OpenAIWSIngressModeOff,
			"openai_apikey_responses_websockets_v2_enabled": false,
		},
	})

	require.NoError(t, err)
	require.NotNil(t, accountRepo.created)
	require.Equal(t, OpenAIWSIngressModeCtxPool, accountRepo.created.Extra["openai_apikey_responses_websockets_v2_mode"])
	require.Equal(t, true, accountRepo.created.Extra["openai_apikey_responses_websockets_v2_enabled"])
	require.Equal(t, []accountCreateDefaultsBindCall{
		{accountID: 900, groupIDs: []int64{4, 5, 12}},
	}, accountRepo.bindCalls)
}

func TestAdminServiceCreateAccountOpenAIAPIKeyKeepsPlatformDefaultGroup(t *testing.T) {
	accountRepo := &accountRepoStubForCreateAccountDefaults{}
	groupRepo := &groupRepoStubForCreateAccountDefaults{
		activeByPlatform: map[string][]Group{
			PlatformOpenAI: {{ID: 3, Name: PlatformOpenAI + "-default", Platform: PlatformOpenAI, Status: StatusActive}},
		},
		active: []Group{{ID: 12, Name: openAIAPIKeyDefaultGroupName, Platform: "mixed", Status: StatusActive}},
	}
	svc := &adminServiceImpl{accountRepo: accountRepo, groupRepo: groupRepo}

	_, err := svc.CreateAccount(context.Background(), &CreateAccountInput{
		Name:     "openai-apikey",
		Platform: PlatformOpenAI,
		Type:     AccountTypeAPIKey,
	})

	require.NoError(t, err)
	require.NotNil(t, accountRepo.created)
	require.Equal(t, OpenAIWSIngressModeCtxPool, accountRepo.created.Extra["openai_apikey_responses_websockets_v2_mode"])
	require.Equal(t, true, accountRepo.created.Extra["openai_apikey_responses_websockets_v2_enabled"])
	require.Equal(t, []accountCreateDefaultsBindCall{
		{accountID: 900, groupIDs: []int64{3, 12}},
	}, accountRepo.bindCalls)
}

func TestAdminServiceCreateAccountOpenAIAPIKeyPreservesExplicitPassthroughMode(t *testing.T) {
	accountRepo := &accountRepoStubForCreateAccountDefaults{}
	groupRepo := &groupRepoStubForCreateAccountDefaults{
		active: []Group{{ID: 12, Name: openAIAPIKeyDefaultGroupName, Platform: "mixed", Status: StatusActive}},
	}
	svc := &adminServiceImpl{accountRepo: accountRepo, groupRepo: groupRepo}

	_, err := svc.CreateAccount(context.Background(), &CreateAccountInput{
		Name:     "openai-apikey",
		Platform: PlatformOpenAI,
		Type:     AccountTypeAPIKey,
		GroupIDs: []int64{12},
		Extra: map[string]any{
			"openai_apikey_responses_websockets_v2_mode": OpenAIWSIngressModePassthrough,
		},
	})

	require.NoError(t, err)
	require.NotNil(t, accountRepo.created)
	require.Equal(t, OpenAIWSIngressModePassthrough, accountRepo.created.Extra["openai_apikey_responses_websockets_v2_mode"])
	require.Equal(t, true, accountRepo.created.Extra["openai_apikey_responses_websockets_v2_enabled"])
	require.Equal(t, []accountCreateDefaultsBindCall{
		{accountID: 900, groupIDs: []int64{12}},
	}, accountRepo.bindCalls)
}
