import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'

const routeState = vi.hoisted(() => ({
  query: {
    token: 'embedded-token',
  } as Record<string, unknown>,
}))

const routerPush = vi.hoisted(() => vi.fn())

const appStoreState = vi.hoisted(() => ({
  siteName: 'Bridgemind',
  contactInfo: '',
  publicSettingsLoaded: true,
  cachedPublicSettings: {
    payment_enabled: true,
    site_name: 'Bridgemind',
  },
  fetchPublicSettings: vi.fn(),
  showError: vi.fn(),
  showInfo: vi.fn(),
  showSuccess: vi.fn(),
}))

const authStoreState = vi.hoisted(() => ({
  token: '',
  isAdmin: false,
  user: {
    id: 1,
    email: 'sam@example.com',
  },
}))

const toDataURL = vi.hoisted(() => vi.fn().mockResolvedValue('data:image/png;base64,qr'))

vi.mock('vue-router', async () => {
  const actual = await vi.importActual<typeof import('vue-router')>('vue-router')
  return {
    ...actual,
    useRoute: () => routeState,
    useRouter: () => ({
      push: routerPush,
    }),
  }
})

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key,
    }),
  }
})

vi.mock('@/stores', () => ({
  useAppStore: () => appStoreState,
}))

vi.mock('@/stores/auth', () => ({
  useAuthStore: () => authStoreState,
}))

vi.mock('qrcode', () => ({
  default: {
    toDataURL,
  },
  toDataURL,
}))

import RechargeCenterView from '../RechargeCenterView.vue'

function createJsonResponse(body: unknown) {
  return {
    ok: true,
    json: vi.fn().mockResolvedValue(body),
  } as Response
}

function createHtmlChallengeResponse() {
  return {
    ok: true,
    json: vi.fn().mockRejectedValue(new SyntaxError('Unexpected token < in JSON at position 0')),
  } as Response
}

function createErrorResponse(status: number, body: unknown) {
  return {
    ok: false,
    status,
    json: vi.fn().mockResolvedValue(body),
  } as Response
}

describe('RechargeCenterView', () => {
  beforeEach(() => {
    routerPush.mockReset()
    appStoreState.fetchPublicSettings.mockReset()
    appStoreState.showError.mockReset()
    appStoreState.showInfo.mockReset()
    appStoreState.showSuccess.mockReset()
    toDataURL.mockClear()

    global.fetch = vi.fn((input: RequestInfo | URL) => {
      const url = String(input)

      if (url === '/api/v1/payment/checkout-info') {
        return Promise.resolve(createJsonResponse({
          code: 0,
          message: 'success',
          data: {
            methods: {
              alipay: {
                daily_limit: 0,
                daily_used: 0,
                daily_remaining: 0,
                single_min: 0,
                single_max: 0,
                fee_rate: 0,
                available: true,
              },
            },
            plans: [
              {
                id: 1,
                group_id: 2,
                group_platform: 'openai',
                group_name: 'openai',
                rate_multiplier: 1,
                daily_limit_usd: 0,
                weekly_limit_usd: 0,
                monthly_limit_usd: 0,
                supported_model_scopes: ['claude'],
                name: '体验版 7天',
                description: 'Bridgemind 体验订阅',
                price: 9.9,
                original_price: 19.9,
                validity_days: 7,
                validity_unit: 'day',
                features: ['OpenAI / Claude / Gemini 通用'],
                product_name: '体验版 7天',
                external_goods_key: 'phte4a',
              },
            ],
          },
        }))
      }

      if (url === '/api/v1/payment/recharge-shop/info') {
        return Promise.resolve(createJsonResponse({
          code: 1,
          msg: 'success',
          data: {
            create_time: 1710000000,
            description: 'merchant info',
            token: 'HQP8RZ4F',
            auth_status: 2,
            contact_mobile: '18017020368',
            contact_qq: '',
            contact_wechat: '',
            nickname: 'Bridgemind 店铺',
          },
        }))
      }

      if (
        url === '/api/v1/payment/recharge-shop/channels' ||
        url === '/api/v1/payment/recharge-shop/categories'
      ) {
        return Promise.resolve(createHtmlChallengeResponse())
      }

      if (url === '/api/v1/payment/orders') {
        return Promise.resolve(createErrorResponse(503, {
          code: 503,
          message: 'payment gateway requires interactive verification',
          reason: 'PAYMENT_GATEWAY_VERIFICATION_REQUIRED',
          metadata: {
            provider: 'ldxpaybridge',
          },
        }))
      }

      throw new Error(`Unexpected fetch: ${url}`)
    }) as typeof fetch
  })

  it('keeps subscription goods visible when storefront APIs are blocked and does not render merchant mobile numbers', async () => {
    const wrapper = mount(RechargeCenterView, {
      global: {
        stubs: {
          Icon: true,
        },
      },
    })

    await flushPromises()
    await flushPromises()
    await flushPromises()

    expect(wrapper.text()).toContain('订阅')
    expect(wrapper.text()).toContain('体验版 7天')
    expect(wrapper.text()).toContain('[推荐]支付宝')
    expect(wrapper.text()).not.toContain('18017020368')
    expect(appStoreState.showError).not.toHaveBeenCalled()
  })

  it('opens the external shop automatically when internal subscription payment requires interactive verification', async () => {
    const openSpy = vi.spyOn(window, 'open').mockImplementation(() => ({ opener: null } as Window))

    try {
      const wrapper = mount(RechargeCenterView, {
        global: {
          stubs: {
            Icon: true,
          },
        },
      })

      await flushPromises()
      await flushPromises()
      await flushPromises()

      await wrapper.get('button.check_pay').trigger('click')
      await flushPromises()

      expect(openSpy).toHaveBeenCalledWith(
        'https://pay.ldxp.cn/item/phte4a',
        '_blank',
        'noopener,noreferrer',
      )
      expect(appStoreState.showInfo).toHaveBeenCalledWith(
        '当前支付通道需要在 pay.ldxp.cn 完成真人验证，已为你打开外部商品页，请在新页面完成验证和支付。',
      )
      expect(appStoreState.showError).not.toHaveBeenCalled()
    } finally {
      openSpy.mockRestore()
    }
  })
})
