package service

import "testing"

func TestAccountIsTLSFingerprintEnabledSupportsOpenAIOAuth(t *testing.T) {
	account := &Account{
		Platform: PlatformOpenAI,
		Type:     AccountTypeOAuth,
		Extra: map[string]any{
			"enable_tls_fingerprint": true,
		},
	}

	if !account.IsTLSFingerprintEnabled() {
		t.Fatal("expected OpenAI OAuth account to support TLS fingerprint")
	}
}
