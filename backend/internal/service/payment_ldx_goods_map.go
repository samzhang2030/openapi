package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/Wei-Shaw/sub2api/ent/paymentproviderinstance"
	"github.com/Wei-Shaw/sub2api/internal/payment"
)

func (s *PaymentConfigService) getEnabledLdxPayBridgeConfig(ctx context.Context) map[string]string {
	if s == nil || s.entClient == nil {
		return nil
	}

	instances, err := s.entClient.PaymentProviderInstance.Query().
		Where(
			paymentproviderinstance.EnabledEQ(true),
			paymentproviderinstance.ProviderKeyEQ(payment.TypeLdxPayBridge),
		).
		Order(paymentproviderinstance.BySortOrder()).
		All(ctx)
	if err != nil {
		return nil
	}

	for _, inst := range instances {
		if !providerSupportsVisibleMethod(inst, payment.TypeAlipay) {
			continue
		}
		cfg, err := s.decryptConfig(inst.Config)
		if err != nil {
			continue
		}
		return cloneStringMap(cfg)
	}

	return nil
}

// GetLdxPayBridgeConfig returns the first enabled ldxpaybridge configuration
// that is available for the visible alipay method.
func (s *PaymentConfigService) GetLdxPayBridgeConfig(ctx context.Context) map[string]string {
	return s.getEnabledLdxPayBridgeConfig(ctx)
}

// GetLdxPayBridgePlanGoodsMap returns the first enabled ldxpaybridge plan-to-goods
// mapping that is available for the visible alipay method.
func (s *PaymentConfigService) GetLdxPayBridgePlanGoodsMap(ctx context.Context) map[int64]string {
	if s == nil || s.entClient == nil {
		return nil
	}

	instances, err := s.entClient.PaymentProviderInstance.Query().
		Where(
			paymentproviderinstance.EnabledEQ(true),
			paymentproviderinstance.ProviderKeyEQ(payment.TypeLdxPayBridge),
		).
		Order(paymentproviderinstance.BySortOrder()).
		All(ctx)
	if err != nil {
		return nil
	}

	for _, inst := range instances {
		if !providerSupportsVisibleMethod(inst, payment.TypeAlipay) {
			continue
		}
		cfg, err := s.decryptConfig(inst.Config)
		if err != nil || len(cfg) == 0 {
			continue
		}
		parsed, err := parseLdxPlanGoodsMapConfig(cfg["planGoodsMap"])
		if err != nil || len(parsed) == 0 {
			continue
		}
		return parsed
	}

	return nil
}

func parseLdxPlanGoodsMapConfig(raw string) (map[int64]string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
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
				return nil, err
			}
			goodsKey = strings.TrimSpace(goodsKey)
			if goodsKey == "" {
				continue
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
			return nil, err
		}
		goodsKey := strings.TrimSpace(value)
		if goodsKey == "" {
			continue
		}
		parsed[planID] = goodsKey
	}
	if len(parsed) == 0 {
		return nil, fmt.Errorf("planGoodsMap is empty")
	}
	return parsed, nil
}

func cloneStringMap(src map[string]string) map[string]string {
	if len(src) == 0 {
		return nil
	}
	dst := make(map[string]string, len(src))
	for key, value := range src {
		dst[key] = value
	}
	return dst
}
