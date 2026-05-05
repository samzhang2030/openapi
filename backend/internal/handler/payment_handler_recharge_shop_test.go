//go:build unit

package handler

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/enttest"
	"github.com/Wei-Shaw/sub2api/internal/payment"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "modernc.org/sqlite"
)

func TestProxyRechargeShopInjectsConfiguredShopToken(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	const shopToken = "HQP8RZ4F"

	var capturedBody string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/shopApi/Shop/info", r.URL.Path)
		require.Equal(t, http.MethodPost, r.Method)

		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		capturedBody = string(body)

		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"code":1,"msg":"success","data":{"nickname":"智桥网关","token":"HQP8RZ4F"}}`)
	}))
	defer upstream.Close()

	configSvc := newRechargeShopPaymentConfigService(t, map[string]string{
		"shopToken": shopToken,
		"apiBase":   upstream.URL,
	})
	handler := NewPaymentHandler(nil, configSvc, nil)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Params = gin.Params{{Key: "action", Value: "info"}}
	ctx.Request = httptest.NewRequest(
		http.MethodPost,
		"/api/v1/payment/recharge-shop/info",
		bytes.NewBufferString(`{}`),
	)
	ctx.Request.Header.Set("Content-Type", "application/json")

	handler.ProxyRechargeShop(ctx)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.JSONEq(t, `{"code":1,"msg":"success","data":{"nickname":"智桥网关","token":"HQP8RZ4F"}}`, recorder.Body.String())
	require.Contains(t, capturedBody, `"token":"`+shopToken+`"`)
}

func TestProxyRechargeShopRejectsUnsupportedAction(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	handler := NewPaymentHandler(nil, &service.PaymentConfigService{}, nil)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Params = gin.Params{{Key: "action", Value: "unsupported"}}
	ctx.Request = httptest.NewRequest(http.MethodPost, "/api/v1/payment/recharge-shop/unsupported", strings.NewReader(`{}`))
	ctx.Request.Header.Set("Content-Type", "application/json")

	handler.ProxyRechargeShop(ctx)

	require.Equal(t, http.StatusBadRequest, recorder.Code)
	require.Contains(t, recorder.Body.String(), "INVALID_RECHARGE_SHOP_ACTION")
}

func newRechargeShopPaymentConfigService(t *testing.T, cfg map[string]string) *service.PaymentConfigService {
	t.Helper()

	dbName := "file:" + strings.NewReplacer("/", "_", " ", "_").Replace(t.Name()) + "?mode=memory&cache=shared"
	db, err := sql.Open("sqlite", dbName)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	_, err = db.Exec("PRAGMA foreign_keys = ON")
	require.NoError(t, err)

	drv := entsql.OpenDB(dialect.SQLite, db)
	client := enttest.NewClient(t, enttest.WithOptions(dbent.Driver(drv)))
	t.Cleanup(func() { _ = client.Close() })

	configSvc := service.NewPaymentConfigService(client, nil, nil)

	encodedConfig, err := json.Marshal(cfg)
	require.NoError(t, err)

	_, err = client.PaymentProviderInstance.Create().
		SetProviderKey(payment.TypeLdxPayBridge).
		SetName("Recharge Shop Proxy").
		SetConfig(string(encodedConfig)).
		SetSupportedTypes("alipay").
		SetEnabled(true).
		SetSortOrder(1).
		Save(context.Background())
	require.NoError(t, err)

	return configSvc
}
