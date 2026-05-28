//go:build unit

package provider

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLdxStorefrontClientPostJSONRawSolvesChallenge(t *testing.T) {
	t.Parallel()

	const challengeArg = "2FD26A8E056988554E4EAB8ADEA6E7FFE999F264"

	var requestCount int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		if requestCount == 1 {
			w.Header().Set("Content-Type", "text/html")
			_, _ = io.WriteString(w, "<html><script>var arg1='"+challengeArg+"';</script></html>")
			return
		}

		if got := r.Header.Get("Content-Type"); !strings.Contains(got, "application/json") {
			t.Fatalf("unexpected content-type %q", got)
		}

		cookie, err := r.Cookie("acw_sc__v2")
		if err != nil {
			t.Fatalf("expected solved challenge cookie, got error: %v", err)
		}
		if cookie.Value != "69eaf24250ca2abedeb36fbdbafd9e8261569cec" {
			t.Fatalf("unexpected challenge cookie %q", cookie.Value)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"code":1,"msg":"success","data":{"nickname":"智桥网关","token":"HQP8RZ4F"}}`)
	}))
	defer server.Close()

	client, err := NewLdxStorefrontClient(map[string]string{
		"shopToken": "HQP8RZ4F",
		"apiBase":   server.URL,
	})
	if err != nil {
		t.Fatalf("NewLdxStorefrontClient returned error: %v", err)
	}

	body, err := client.PostJSONRaw(context.Background(), "/shopApi/Shop/info", map[string]any{
		"token": "HQP8RZ4F",
	})
	if err != nil {
		t.Fatalf("PostJSONRaw returned error: %v", err)
	}
	if !strings.Contains(string(body), `"nickname":"智桥网关"`) {
		t.Fatalf("unexpected response body: %s", string(body))
	}
	if requestCount != 2 {
		t.Fatalf("requestCount = %d, want 2", requestCount)
	}
}

func TestLdxStorefrontClientPostJSONRawDetectsInteractiveVerification(t *testing.T) {
	t.Parallel()

	const traceID = "76fdacae17771905698388471e"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = io.WriteString(w, `<!doctypehtml>
<script src="//o.alicdn.com/captcha-frontend/aliyunCaptcha/AliyunCaptcha.js"></script>
<textarea id="renderData" style="display:none">var requestInfo = {"traceid":"`+traceID+`","sceneId":"19x5u7lo"};</textarea>`)
	}))
	defer server.Close()

	client, err := NewLdxStorefrontClient(map[string]string{
		"shopToken": "HQP8RZ4F",
		"apiBase":   server.URL,
	})
	if err != nil {
		t.Fatalf("NewLdxStorefrontClient returned error: %v", err)
	}

	_, err = client.PostJSONRaw(context.Background(), "/shopApi/Shop/getGoodsPrice", map[string]any{
		"goods_key":  "y8virm",
		"quantity":   1,
		"channel_id": 1,
	})
	if err == nil {
		t.Fatal("expected interactive verification error, got nil")
	}

	var verificationErr *LdxVerificationRequiredError
	if !errors.As(err, &verificationErr) {
		t.Fatalf("expected LdxVerificationRequiredError, got %T: %v", err, err)
	}
	if verificationErr.Path != "/shopApi/Shop/getGoodsPrice" {
		t.Fatalf("verification path = %q, want %q", verificationErr.Path, "/shopApi/Shop/getGoodsPrice")
	}
	if verificationErr.TraceID != traceID {
		t.Fatalf("verification trace_id = %q, want %q", verificationErr.TraceID, traceID)
	}
}
