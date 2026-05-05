//go:build unit

package service

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/payment"
)

func TestParseLdxPlanGoodsMapConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		raw     string
		wantLen int
		want    string
		wantErr bool
	}{
		{name: "empty", raw: "", wantLen: 0},
		{name: "csv", raw: "1=phte4a,2=y8virm,3=ounq0d", wantLen: 3, want: "y8virm"},
		{name: "json", raw: `{"1":"phte4a","2":"y8virm"}`, wantLen: 2, want: "y8virm"},
		{name: "invalid", raw: "bad", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := parseLdxPlanGoodsMapConfig(tt.raw)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("parseLdxPlanGoodsMapConfig returned error: %v", err)
			}
			if len(got) != tt.wantLen {
				t.Fatalf("len(got) = %d, want %d", len(got), tt.wantLen)
			}
			if tt.want != "" && got[2] != tt.want {
				t.Fatalf("got[2] = %q, want %q", got[2], tt.want)
			}
		})
	}
}

func TestGetLdxPayBridgePlanGoodsMap(t *testing.T) {
	t.Parallel()

	client := newPaymentConfigServiceTestClient(t)
	svc := &PaymentConfigService{entClient: client}
	ctx := context.Background()

	encValid, err := svc.encryptConfig(map[string]string{
		"shopToken":    "HQP8RZ4F",
		"planGoodsMap": "1=phte4a,2=y8virm,3=ounq0d",
	})
	if err != nil {
		t.Fatalf("encrypt valid config: %v", err)
	}
	_, err = client.PaymentProviderInstance.Create().
		SetProviderKey(payment.TypeLdxPayBridge).
		SetName("valid").
		SetConfig(encValid).
		SetSupportedTypes("alipay").
		SetEnabled(true).
		SetSortOrder(1).
		Save(ctx)
	if err != nil {
		t.Fatalf("create valid instance: %v", err)
	}

	got := svc.GetLdxPayBridgePlanGoodsMap(ctx)
	if got[1] != "phte4a" || got[2] != "y8virm" || got[3] != "ounq0d" {
		t.Fatalf("unexpected map: %#v", got)
	}
}
