//go:build unit

package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

type accountRepoStubForAccountServiceCreate struct {
	accountRepoStub
	created *Account
}

func (s *accountRepoStubForAccountServiceCreate) Create(_ context.Context, account *Account) error {
	account.ID = 101
	clone := *account
	s.created = &clone
	return nil
}

func TestAccountServiceCreateDefaultsSchedulable(t *testing.T) {
	repo := &accountRepoStubForAccountServiceCreate{}
	svc := &AccountService{accountRepo: repo}

	account, err := svc.Create(context.Background(), CreateAccountRequest{
		Name:        "new-api-key",
		Platform:    PlatformAnthropic,
		Type:        AccountTypeAPIKey,
		Credentials: map[string]any{"api_key": "sk-test"},
	})

	require.NoError(t, err)
	require.NotNil(t, account)
	require.NotNil(t, repo.created)
	require.True(t, repo.created.Schedulable)
	require.True(t, account.Schedulable)
}
