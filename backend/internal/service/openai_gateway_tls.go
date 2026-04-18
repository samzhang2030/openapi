package service

import "github.com/Wei-Shaw/sub2api/internal/pkg/tlsfingerprint"

// resolveTLSProfile resolves the runtime TLS fingerprint profile for OpenAI
// upstream requests. When the profile service is unavailable but the account
// explicitly enables TLS fingerprinting, fall back to the built-in default so
// OAuth forwarding still uses the Node.js-style handshake.
func (s *OpenAIGatewayService) resolveTLSProfile(account *Account) *tlsfingerprint.Profile {
	if account == nil || !account.IsTLSFingerprintEnabled() {
		return nil
	}
	if s != nil && s.tlsFPProfileService != nil {
		if profile := s.tlsFPProfileService.ResolveTLSProfile(account); profile != nil {
			return profile
		}
	}
	return &tlsfingerprint.Profile{Name: "Built-in Default (Node.js 24.x)"}
}
