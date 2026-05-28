package provider

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/payment"
)

const (
	ldxPayBridgeDefaultAPIBase = "https://pay.ldxp.cn"
	ldxPayBridgeHTTPTimeout    = 15 * time.Second
	ldxPayBridgeResponseLimit  = 2 << 20
	ldxPayBridgeRetryLimit     = 3
	ldxPayBridgeXORKey         = "3000176000856006061501533003690027800375"
)

var ldxPayBridgeReorder = [...]int{
	0xf, 0x23, 0x1d, 0x18, 0x21, 0x10, 0x1, 0x26, 0xa, 0x9,
	0x13, 0x1f, 0x28, 0x1b, 0x16, 0x17, 0x19, 0xd, 0x6, 0xb,
	0x27, 0x12, 0x14, 0x8, 0xe, 0x15, 0x20, 0x1a, 0x2, 0x1e,
	0x7, 0x4, 0x11, 0x5, 0x3, 0x1c, 0x22, 0x25, 0xc, 0x24,
}

type ldxPayBridge struct {
	instanceID   string
	config       map[string]string
	planGoodsMap map[int64]string
	client       *LdxStorefrontClient
}

type ldxResponse[T any] struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data T      `json:"data"`
}

type ldxChannel struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	PayType struct {
		Name string `json:"name"`
	} `json:"paytype"`
}

type ldxGoodsPriceData struct {
	TotalAmount    any `json:"total_amount"`
	OriginalAmount any `json:"original_amount"`
}

type ldxOrderData struct {
	TradeNo     string `json:"trade_no"`
	PayURL      string `json:"payurl"`
	TotalAmount any    `json:"total_amount"`
}

// NewLdxPayBridge creates a pay.ldxp.cn compatibility bridge for internal subscription orders.
func NewLdxPayBridge(instanceID string, config map[string]string) (*ldxPayBridge, error) {
	planGoodsMap, err := parseLdxPlanGoodsMap(config["planGoodsMap"])
	if err != nil {
		return nil, fmt.Errorf("ldxpaybridge invalid planGoodsMap: %w", err)
	}

	client, err := NewLdxStorefrontClient(config)
	if err != nil {
		return nil, err
	}

	return &ldxPayBridge{
		instanceID:   instanceID,
		config:       config,
		planGoodsMap: planGoodsMap,
		client:       client,
	}, nil
}

func (p *ldxPayBridge) Name() string        { return "LdxPay Bridge" }
func (p *ldxPayBridge) ProviderKey() string { return payment.TypeLdxPayBridge }
func (p *ldxPayBridge) SupportedTypes() []payment.PaymentType {
	return []payment.PaymentType{payment.TypeAlipay}
}

func (p *ldxPayBridge) MerchantIdentityMetadata() map[string]string {
	if p == nil || p.client == nil || p.client.ShopToken() == "" {
		return nil
	}
	return map[string]string{"shop_token": p.client.ShopToken()}
}

func (p *ldxPayBridge) CreatePayment(ctx context.Context, req payment.CreatePaymentRequest) (*payment.CreatePaymentResponse, error) {
	if payment.GetBasePaymentType(req.PaymentType) != payment.TypeAlipay {
		return nil, fmt.Errorf("ldxpaybridge only supports alipay")
	}
	if strings.TrimSpace(req.OrderType) != payment.OrderTypeSubscription {
		return nil, fmt.Errorf("ldxpaybridge only supports subscription orders")
	}
	if req.PlanID <= 0 {
		return nil, fmt.Errorf("ldxpaybridge requires a valid plan id")
	}
	goodsKey := strings.TrimSpace(p.planGoodsMap[req.PlanID])
	if goodsKey == "" {
		return nil, fmt.Errorf("ldxpaybridge missing goods mapping for plan %d", req.PlanID)
	}
	contact := strings.TrimSpace(req.Contact)
	if contact == "" {
		return nil, fmt.Errorf("ldxpaybridge requires a contact value")
	}

	channelID, err := p.client.ResolveChannelID(ctx)
	if err != nil {
		return nil, err
	}

	var priceResp ldxResponse[ldxGoodsPriceData]
	if err := p.client.PostJSON(ctx, "/shopApi/Shop/getGoodsPrice", map[string]any{
		"goods_key":  goodsKey,
		"quantity":   1,
		"channel_id": channelID,
	}, &priceResp); err != nil {
		return nil, err
	}
	if priceResp.Code != 1 {
		return nil, fmt.Errorf("ldxpaybridge getGoodsPrice failed: %s", strings.TrimSpace(priceResp.Msg))
	}
	quotedAmount, err := ldxFloatFromAny(priceResp.Data.TotalAmount)
	if err != nil {
		return nil, fmt.Errorf("ldxpaybridge parse goods price: %w", err)
	}
	expectedAmount, err := strconv.ParseFloat(strings.TrimSpace(req.Amount), 64)
	if err != nil {
		return nil, fmt.Errorf("ldxpaybridge invalid request amount %q: %w", req.Amount, err)
	}
	if math.Abs(quotedAmount-expectedAmount) > 0.009 {
		return nil, fmt.Errorf(
			"ldxpaybridge amount mismatch for plan %d: goods %.2f, order %.2f",
			req.PlanID,
			quotedAmount,
			expectedAmount,
		)
	}

	var orderResp ldxResponse[ldxOrderData]
	if err := p.client.PostJSON(ctx, "/shopApi/Pay/order", map[string]any{
		"goods_key":        goodsKey,
		"quantity":         1,
		"channel_id":       channelID,
		"contact":          contact,
		"coupon_code":      "",
		"query_password":   "",
		"select_cards_ids": []string{},
		"extend":           map[string]any{},
	}, &orderResp); err != nil {
		return nil, err
	}
	if orderResp.Code != 1 {
		return nil, fmt.Errorf("ldxpaybridge order failed: %s", strings.TrimSpace(orderResp.Msg))
	}
	if strings.TrimSpace(orderResp.Data.TradeNo) == "" || strings.TrimSpace(orderResp.Data.PayURL) == "" {
		return nil, fmt.Errorf("ldxpaybridge order returned incomplete payment data")
	}

	return &payment.CreatePaymentResponse{
		TradeNo: orderResp.Data.TradeNo,
		PayURL:  orderResp.Data.PayURL,
	}, nil
}

func (p *ldxPayBridge) QueryOrder(ctx context.Context, tradeNo string) (*payment.QueryOrderResponse, error) {
	tradeNo = strings.TrimSpace(tradeNo)
	if tradeNo == "" {
		return nil, fmt.Errorf("ldxpaybridge query trade no is required")
	}

	var queryResp ldxResponse[map[string]any]
	if err := p.client.PostJSON(ctx, "/shopApi/Pay/query", map[string]any{"trade_no": tradeNo}, &queryResp); err != nil {
		return nil, err
	}

	status := payment.ProviderStatusFailed
	switch {
	case queryResp.Code == 1:
		status = payment.ProviderStatusPaid
	case queryResp.Code == 0 && strings.Contains(strings.ToLower(strings.TrimSpace(queryResp.Msg)), "not pay"):
		status = payment.ProviderStatusPending
	default:
		return nil, fmt.Errorf("ldxpaybridge query failed: %s", strings.TrimSpace(queryResp.Msg))
	}

	amount, _ := ldxFloatFromAny(ldxFirstAmountField(queryResp.Data, "total_amount", "amount", "money"))
	return &payment.QueryOrderResponse{
		TradeNo:  tradeNo,
		Status:   status,
		Amount:   amount,
		Metadata: p.MerchantIdentityMetadata(),
	}, nil
}

func (p *ldxPayBridge) VerifyNotification(context.Context, string, map[string]string) (*payment.PaymentNotification, error) {
	return nil, nil
}

func (p *ldxPayBridge) Refund(context.Context, payment.RefundRequest) (*payment.RefundResponse, error) {
	return nil, fmt.Errorf("ldxpaybridge does not support refunds")
}

func parseLdxPlanGoodsMap(raw string) (map[int64]string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, fmt.Errorf("planGoodsMap is required")
	}

	parsed := make(map[int64]string)
	if strings.HasPrefix(raw, "{") {
		var jsonMap map[string]string
		if err := json.Unmarshal([]byte(raw), &jsonMap); err != nil {
			return nil, err
		}
		for key, goodsKey := range jsonMap {
			planID, err := strconv.ParseInt(strings.TrimSpace(key), 10, 64)
			if err != nil || planID <= 0 {
				return nil, fmt.Errorf("invalid plan id %q", key)
			}
			goodsKey = strings.TrimSpace(goodsKey)
			if goodsKey == "" {
				return nil, fmt.Errorf("plan %d has an empty goods key", planID)
			}
			parsed[planID] = goodsKey
		}
		return parsed, nil
	}

	parts := strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == ';' || r == '\n' || r == '\r'
	})
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		key, value, ok := strings.Cut(part, "=")
		if !ok {
			return nil, fmt.Errorf("invalid entry %q", part)
		}
		planID, err := strconv.ParseInt(strings.TrimSpace(key), 10, 64)
		if err != nil || planID <= 0 {
			return nil, fmt.Errorf("invalid plan id %q", key)
		}
		goodsKey := strings.TrimSpace(value)
		if goodsKey == "" {
			return nil, fmt.Errorf("plan %d has an empty goods key", planID)
		}
		parsed[planID] = goodsKey
	}

	if len(parsed) == 0 {
		return nil, fmt.Errorf("planGoodsMap is empty")
	}
	return parsed, nil
}

func chooseLdxChannel(channels []ldxChannel) int64 {
	var fallback int64
	for _, channel := range channels {
		if channel.ID <= 0 {
			continue
		}
		if fallback == 0 {
			fallback = channel.ID
		}
		candidate := strings.ToLower(strings.TrimSpace(channel.Name + " " + channel.PayType.Name))
		if strings.Contains(candidate, "alipay") || strings.Contains(candidate, "\u652f\u4ed8\u5b9d") {
			return channel.ID
		}
	}
	return fallback
}

func extractLdxChallengeArg(body []byte) string {
	text := string(body)
	start := strings.Index(text, "var arg1='")
	if start < 0 {
		return ""
	}
	start += len("var arg1='")
	end := strings.Index(text[start:], "'")
	if end < 0 {
		return ""
	}
	return strings.TrimSpace(text[start : start+end])
}

func solveLdxChallengeCookie(arg1 string) (string, error) {
	arg1 = strings.TrimSpace(arg1)
	if len(arg1) != len(ldxPayBridgeReorder) {
		return "", fmt.Errorf("unexpected acw challenge length %d", len(arg1))
	}

	reordered := make([]byte, len(ldxPayBridgeReorder))
	for idx := 0; idx < len(arg1); idx++ {
		for slot, order := range ldxPayBridgeReorder {
			if order == idx+1 {
				reordered[slot] = arg1[idx]
				break
			}
		}
	}

	reorderedBytes, err := hex.DecodeString(string(reordered))
	if err != nil {
		return "", fmt.Errorf("decode reordered challenge: %w", err)
	}
	keyBytes, err := hex.DecodeString(ldxPayBridgeXORKey)
	if err != nil {
		return "", fmt.Errorf("decode xor key: %w", err)
	}
	if len(reorderedBytes) != len(keyBytes) {
		return "", fmt.Errorf("challenge length mismatch")
	}

	result := make([]byte, len(reorderedBytes))
	for idx := range reorderedBytes {
		result[idx] = reorderedBytes[idx] ^ keyBytes[idx]
	}
	return hex.EncodeToString(result), nil
}

func ldxFloatFromAny(value any) (float64, error) {
	switch typed := value.(type) {
	case nil:
		return 0, nil
	case float64:
		return typed, nil
	case float32:
		return float64(typed), nil
	case int:
		return float64(typed), nil
	case int64:
		return float64(typed), nil
	case json.Number:
		return typed.Float64()
	case string:
		if strings.TrimSpace(typed) == "" {
			return 0, nil
		}
		return strconv.ParseFloat(strings.TrimSpace(typed), 64)
	default:
		return 0, fmt.Errorf("unsupported amount type %T", value)
	}
}

func ldxFirstAmountField(data map[string]any, keys ...string) any {
	for _, key := range keys {
		if value, ok := data[key]; ok {
			return value
		}
	}
	return nil
}

var (
	_ payment.Provider                 = (*ldxPayBridge)(nil)
	_ payment.MerchantIdentityProvider = (*ldxPayBridge)(nil)
)
