package service

import (
	"net/http"

	"github.com/Wei-Shaw/sub2api/internal/pkg/tlsfingerprint"
)

func (s *OpenAIGatewayService) resolveTLSProfile(account *Account) *tlsfingerprint.Profile {
	if account == nil || !account.IsTLSFingerprintEnabled() {
		return nil
	}
	if s != nil && s.tlsFPProfileService != nil {
		return s.tlsFPProfileService.ResolveTLSProfile(account)
	}
	return &tlsfingerprint.Profile{Name: "Built-in Default (Node.js 24.x)"}
}

func (s *OpenAIGatewayService) doUpstreamRequest(req *http.Request, proxyURL string, account *Account) (*http.Response, error) {
	profile := s.resolveTLSProfile(account)
	if profile != nil {
		return s.httpUpstream.DoWithTLS(req, proxyURL, account.ID, account.Concurrency, profile)
	}
	return s.httpUpstream.Do(req, proxyURL, account.ID, account.Concurrency)
}
