package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// LdxStorefrontClient proxies pay.ldxp.cn storefront APIs through the backend so
// the browser does not need to talk to the upstream shop directly.
type LdxStorefrontClient struct {
	apiBase          string
	shopToken        string
	defaultChannelID int64
	httpClient       *http.Client
}

// LdxVerificationRequiredError indicates the upstream storefront requested
// an interactive anti-bot verification page instead of the expected API JSON.
type LdxVerificationRequiredError struct {
	Path    string
	TraceID string
}

func (e *LdxVerificationRequiredError) Error() string {
	if e == nil {
		return "ldxpaybridge upstream interactive verification required"
	}
	path := strings.TrimSpace(e.Path)
	traceID := strings.TrimSpace(e.TraceID)
	switch {
	case path != "" && traceID != "":
		return fmt.Sprintf("ldxpaybridge upstream interactive verification required for %s (trace_id=%s)", path, traceID)
	case path != "":
		return fmt.Sprintf("ldxpaybridge upstream interactive verification required for %s", path)
	case traceID != "":
		return fmt.Sprintf("ldxpaybridge upstream interactive verification required (trace_id=%s)", traceID)
	default:
		return "ldxpaybridge upstream interactive verification required"
	}
}

// NewLdxStorefrontClient creates a storefront client from provider config.
func NewLdxStorefrontClient(config map[string]string) (*LdxStorefrontClient, error) {
	shopToken := strings.TrimSpace(config["shopToken"])
	if shopToken == "" {
		return nil, fmt.Errorf("ldxpaybridge config missing required key: shopToken")
	}

	apiBase := strings.TrimSpace(config["apiBase"])
	if apiBase == "" {
		apiBase = ldxPayBridgeDefaultAPIBase
	}
	apiBase = strings.TrimRight(apiBase, "/")

	var (
		defaultChannelID int64
		err              error
	)
	if raw := strings.TrimSpace(config["defaultChannelId"]); raw != "" {
		defaultChannelID, err = strconv.ParseInt(raw, 10, 64)
		if err != nil || defaultChannelID <= 0 {
			return nil, fmt.Errorf("ldxpaybridge invalid defaultChannelId: %s", raw)
		}
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("ldxpaybridge init cookie jar: %w", err)
	}

	return &LdxStorefrontClient{
		apiBase:          apiBase,
		shopToken:        shopToken,
		defaultChannelID: defaultChannelID,
		httpClient: &http.Client{
			Timeout: ldxPayBridgeHTTPTimeout,
			Jar:     jar,
		},
	}, nil
}

func (c *LdxStorefrontClient) ShopToken() string {
	if c == nil {
		return ""
	}
	return c.shopToken
}

func (c *LdxStorefrontClient) DefaultChannelID() int64 {
	if c == nil {
		return 0
	}
	return c.defaultChannelID
}

func (c *LdxStorefrontClient) PostJSON(ctx context.Context, path string, payload any, out any) error {
	body, err := c.PostJSONRaw(ctx, path, payload)
	if err != nil {
		return err
	}
	if out == nil {
		return nil
	}
	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("ldxpaybridge parse %s: %w", path, err)
	}
	return nil
}

func (c *LdxStorefrontClient) PostJSONRaw(ctx context.Context, path string, payload any) ([]byte, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("ldxpaybridge marshal %s: %w", path, err)
	}
	endpoint := c.apiBase + path

	for attempt := 0; attempt < ldxPayBridgeRetryLimit; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
		if err != nil {
			return nil, fmt.Errorf("ldxpaybridge create request %s: %w", path, err)
		}
		req.Header.Set("Accept", "application/json, text/plain, */*")
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("ldxpaybridge request %s: %w", path, err)
		}

		respBody, readErr := io.ReadAll(io.LimitReader(resp.Body, ldxPayBridgeResponseLimit))
		_ = resp.Body.Close()
		if readErr != nil {
			return nil, fmt.Errorf("ldxpaybridge read %s: %w", path, readErr)
		}

		if arg1 := extractLdxChallengeArg(respBody); arg1 != "" {
			if err := c.storeChallengeCookie(endpoint, arg1); err != nil {
				return nil, fmt.Errorf("ldxpaybridge solve acw challenge: %w", err)
			}
			continue
		}
		if isLdxInteractiveVerificationPage(respBody) {
			return nil, &LdxVerificationRequiredError{
				Path:    path,
				TraceID: extractLdxVerificationTraceID(respBody),
			}
		}

		if resp.StatusCode >= http.StatusBadRequest {
			return nil, fmt.Errorf("ldxpaybridge %s returned http %d", path, resp.StatusCode)
		}
		return respBody, nil
	}

	return nil, fmt.Errorf("ldxpaybridge acw challenge retry exhausted for %s", path)
}

func (c *LdxStorefrontClient) ResolveChannelID(ctx context.Context) (int64, error) {
	if c.defaultChannelID > 0 {
		return c.defaultChannelID, nil
	}

	var channelResp ldxResponse[[]ldxChannel]
	if err := c.PostJSON(ctx, "/shopApi/Shop/getUserChannel", map[string]any{"token": c.shopToken}, &channelResp); err != nil {
		return 0, err
	}
	if channelResp.Code != 1 {
		return 0, fmt.Errorf("ldxpaybridge getUserChannel failed: %s", strings.TrimSpace(channelResp.Msg))
	}
	channelID := chooseLdxChannel(channelResp.Data)
	if channelID <= 0 {
		return 0, fmt.Errorf("ldxpaybridge could not resolve an alipay channel")
	}
	return channelID, nil
}

func (c *LdxStorefrontClient) storeChallengeCookie(rawURL string, arg1 string) error {
	value, err := solveLdxChallengeCookie(arg1)
	if err != nil {
		return err
	}
	if c.httpClient == nil || c.httpClient.Jar == nil {
		return fmt.Errorf("ldxpaybridge cookie jar is not initialized")
	}
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("parse ldxpaybridge url: %w", err)
	}
	c.httpClient.Jar.SetCookies(parsedURL, []*http.Cookie{{
		Name:    "acw_sc__v2",
		Value:   value,
		Path:    "/",
		Expires: time.Now().Add(time.Hour),
	}})
	return nil
}

func isLdxInteractiveVerificationPage(body []byte) bool {
	text := strings.ToLower(string(body))
	return strings.Contains(text, "aliyuncaptcha") ||
		strings.Contains(text, "cf_app_waf") ||
		strings.Contains(text, "please complete the operation to verify that you are a real person") ||
		strings.Contains(string(body), "请完成以下操作，验证您是真人")
}

func extractLdxVerificationTraceID(body []byte) string {
	text := string(body)
	for _, marker := range []string{`"traceid":"`, `TraceID：`, `TraceID:`} {
		start := strings.Index(text, marker)
		if start < 0 {
			continue
		}
		start += len(marker)
		end := start
		for end < len(text) {
			ch := text[end]
			if (ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F') {
				end++
				continue
			}
			break
		}
		traceID := strings.TrimSpace(text[start:end])
		if traceID != "" {
			return traceID
		}
	}
	return ""
}
