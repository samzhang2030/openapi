//go:build unit

package provider

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/payment"
)

func TestSolveLdxChallengeCookie(t *testing.T) {
	t.Parallel()

	got, err := solveLdxChallengeCookie("2FD26A8E056988554E4EAB8ADEA6E7FFE999F264")
	if err != nil {
		t.Fatalf("solveLdxChallengeCookie returned error: %v", err)
	}
	if got != "69eaf24250ca2abedeb36fbdbafd9e8261569cec" {
		t.Fatalf("solveLdxChallengeCookie = %q", got)
	}
}

func TestParseLdxPlanGoodsMap(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		raw     string
		wantLen int
		wantErr bool
	}{
		{name: "csv format", raw: "1=phte4a,2=y8virm,3=ounq0d", wantLen: 3},
		{name: "json format", raw: `{"1":"phte4a","2":"y8virm"}`, wantLen: 2},
		{name: "invalid entry", raw: "1=ok,bad", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := parseLdxPlanGoodsMap(tt.raw)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("parseLdxPlanGoodsMap returned error: %v", err)
			}
			if len(got) != tt.wantLen {
				t.Fatalf("map length = %d, want %d", len(got), tt.wantLen)
			}
		})
	}
}

func TestLdxPayBridgeCreatePayment(t *testing.T) {
	t.Parallel()

	const challengeArg = "2FD26A8E056988554E4EAB8ADEA6E7FFE999F264"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/shopApi/Shop/getUserChannel":
			if _, err := r.Cookie("acw_sc__v2"); err != nil {
				w.Header().Set("Content-Type", "text/html")
				_, _ = io.WriteString(w, "<html><script>var arg1='"+challengeArg+"';</script></html>")
				return
			}
			if got := r.Header.Get("Content-Type"); !strings.Contains(got, "application/json") {
				t.Fatalf("unexpected content-type %q", got)
			}
			var payload map[string]any
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Fatalf("decode getUserChannel payload: %v", err)
			}
			if payload["token"] != "HQP8RZ4F" {
				t.Fatalf("unexpected token payload: %+v", payload)
			}
			channel := ldxChannel{ID: 1, Name: "\u652f\u4ed8\u5b9d"}
			channel.PayType.Name = "\u652f\u4ed8\u5b9d"
			_ = json.NewEncoder(w).Encode(ldxResponse[[]ldxChannel]{
				Code: 1,
				Data: []ldxChannel{channel},
			})
		case "/shopApi/Shop/getGoodsPrice":
			var payload map[string]any
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Fatalf("decode getGoodsPrice payload: %v", err)
			}
			if payload["goods_key"] != "y8virm" {
				t.Fatalf("unexpected goods_key payload: %+v", payload)
			}
			_ = json.NewEncoder(w).Encode(ldxResponse[ldxGoodsPriceData]{
				Code: 1,
				Data: ldxGoodsPriceData{TotalAmount: "29.90"},
			})
		case "/shopApi/Pay/order":
			var payload map[string]any
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Fatalf("decode order payload: %v", err)
			}
			if payload["contact"] != "aaa@qq.com" {
				t.Fatalf("unexpected contact payload: %+v", payload)
			}
			_ = json.NewEncoder(w).Encode(ldxResponse[ldxOrderData]{
				Code: 1,
				Data: ldxOrderData{
					TradeNo:     "LD2604248QZQYV",
					PayURL:      "https://pay.ldxp.cn/pay/LD2604248QZQYV",
					TotalAmount: "29.90",
				},
			})
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()

	provider, err := NewLdxPayBridge("test-instance", map[string]string{
		"shopToken":    "HQP8RZ4F",
		"planGoodsMap": "2=y8virm",
		"apiBase":      server.URL,
	})
	if err != nil {
		t.Fatalf("NewLdxPayBridge returned error: %v", err)
	}

	resp, err := provider.CreatePayment(context.Background(), payment.CreatePaymentRequest{
		Amount:      "29.90",
		PaymentType: payment.TypeAlipay,
		OrderType:   payment.OrderTypeSubscription,
		PlanID:      2,
		Contact:     "aaa@qq.com",
	})
	if err != nil {
		t.Fatalf("CreatePayment returned error: %v", err)
	}
	if resp.TradeNo != "LD2604248QZQYV" {
		t.Fatalf("trade_no = %q", resp.TradeNo)
	}
	if resp.PayURL != "https://pay.ldxp.cn/pay/LD2604248QZQYV" {
		t.Fatalf("pay_url = %q", resp.PayURL)
	}
}

func TestLdxPayBridgeCreatePaymentRejectsAmountMismatch(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/shopApi/Shop/getGoodsPrice":
			_ = json.NewEncoder(w).Encode(ldxResponse[ldxGoodsPriceData]{
				Code: 1,
				Data: ldxGoodsPriceData{TotalAmount: "39.90"},
			})
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()

	provider, err := NewLdxPayBridge("test-instance", map[string]string{
		"shopToken":        "HQP8RZ4F",
		"planGoodsMap":     "2=y8virm",
		"defaultChannelId": "1",
		"apiBase":          server.URL,
	})
	if err != nil {
		t.Fatalf("NewLdxPayBridge returned error: %v", err)
	}

	_, err = provider.CreatePayment(context.Background(), payment.CreatePaymentRequest{
		Amount:      "29.90",
		PaymentType: payment.TypeAlipay,
		OrderType:   payment.OrderTypeSubscription,
		PlanID:      2,
		Contact:     "aaa@qq.com",
	})
	if err == nil {
		t.Fatal("expected amount mismatch error, got nil")
	}
	if !strings.Contains(err.Error(), "amount mismatch") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLdxPayBridgeQueryOrder(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/shopApi/Pay/query" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}

		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode query payload: %v", err)
		}
		tradeNo, _ := payload["trade_no"].(string)
		if tradeNo == "paid-trade" {
			_ = json.NewEncoder(w).Encode(ldxResponse[map[string]any]{
				Code: 1,
				Data: map[string]any{"total_amount": "79.90"},
			})
			return
		}
		_ = json.NewEncoder(w).Encode(ldxResponse[map[string]any]{
			Code: 0,
			Msg:  "not pay",
			Data: map[string]any{},
		})
	}))
	defer server.Close()

	provider, err := NewLdxPayBridge("test-instance", map[string]string{
		"shopToken":    "HQP8RZ4F",
		"planGoodsMap": "3=ounq0d",
		"apiBase":      server.URL,
	})
	if err != nil {
		t.Fatalf("NewLdxPayBridge returned error: %v", err)
	}

	paidResp, err := provider.QueryOrder(context.Background(), "paid-trade")
	if err != nil {
		t.Fatalf("QueryOrder(paid) returned error: %v", err)
	}
	if paidResp.Status != payment.ProviderStatusPaid || paidResp.Amount != 79.90 {
		t.Fatalf("paid response = %+v", paidResp)
	}

	pendingResp, err := provider.QueryOrder(context.Background(), "pending-trade")
	if err != nil {
		t.Fatalf("QueryOrder(pending) returned error: %v", err)
	}
	if pendingResp.Status != payment.ProviderStatusPending {
		t.Fatalf("pending response = %+v", pendingResp)
	}
}
