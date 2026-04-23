<template>
  <div class="store-page" :class="{ 'store-page-embedded': embeddedMode }">
    <div v-if="floatingActions.length" class="bangz_box" aria-label="快捷入口">
      <a
        v-for="action in floatingActions"
        :key="action.key"
        class="item"
        :class="`item-${action.tone}`"
        :href="action.href"
        target="_blank"
        rel="noopener noreferrer"
      >
        <span>{{ action.label }}</span>
      </a>
    </div>

    <header class="header">
      <div class="container header-inner">
        <div class="header_left">
          <button type="button" class="header_logo" @click="navigateBack">
            <img src="/logo.png" alt="Bridgemind" />
            <span>{{ siteName }}</span>
          </button>

          <button
            v-if="merchantContactText"
            type="button"
            class="online-btn"
            @click="copyMerchantContact"
          >
            <svg class="icon" viewBox="0 0 1024 1024" width="16" height="16" aria-hidden="true">
              <path
                d="M877.714286 170.666667v585.142857H575.463619l-282.526476 196.851809V755.809524H146.285714V170.666667h731.428572z m-97.52381 97.523809H243.809524v390.095238h146.651428l-0.024381 107.568762L544.841143 658.285714H780.190476V268.190476z m-365.714286 121.904762a73.142857 73.142857 0 1 1 0 146.285714 73.142857 73.142857 0 0 1 0-146.285714z m195.04762 0a73.142857 73.142857 0 1 1 0 146.285714 73.142857 73.142857 0 0 1 0-146.285714z"
                fill="#ffffff"
              />
            </svg>
            <span>{{ merchantContactText }}</span>
          </button>
        </div>

        <div class="header_right">
          <a
            v-if="orderQueryUrl"
            class="header_button"
            :href="orderQueryUrl"
            target="_blank"
            rel="noopener noreferrer"
          >
            <svg class="icon" viewBox="0 0 1024 1024" width="22" height="22" aria-hidden="true">
              <path
                d="M455.253333 657.92c117.76 0 213.333333-95.573333 213.333334-213.333333s-95.573333-213.333333-213.333334-213.333334-213.333333 95.573333-213.333333 213.333334 95.573333 213.333333 213.333333 213.333333z m229.76-22.4l169.813334 169.813333c16.64 16.64 16.64 43.733333 0 60.373334-16.64 16.64-43.733333 16.64-60.373334 0l-172.8-172.8c-47.573333 32-104.746667 50.56-166.4 50.56-164.906667 0-298.666667-133.76-298.666666-298.666667s133.76-298.666667 298.666666-298.666667 298.666667 133.76 298.666667 298.666667c0 72.32-25.813333 138.88-68.906667 190.72z"
                fill="#ffffff"
              />
            </svg>
            <span>订单查询</span>
          </a>
        </div>
      </div>
    </header>

    <section class="section details__section section--first section--last">
      <div class="container">
        <div v-if="loading" class="store-loading">
          <div class="loading-dot"></div>
          <span>正在同步店铺数据...</span>
        </div>

        <template v-else>
          <div class="section__title-wrap">
            <h2 class="section__title section__title--pre">商家信息</h2>
          </div>

          <hr class="mt-4" />

          <div class="merchant-meta-row">
            <div v-if="merchantInfo?.auth_status === 2" class="merchant-meta-item">
              <span class="merchant-meta-icon">◆</span>
              <span>商家认证：<strong>已认证</strong></span>
            </div>

            <div v-if="formattedCreateDate" class="merchant-meta-item">
              <span class="merchant-meta-icon">◈</span>
              <span>开店时间：{{ formattedCreateDate }}</span>
            </div>

            <div v-if="merchantContactText" class="merchant-meta-item">
              <span class="merchant-meta-icon">✦</span>
              <span>{{ merchantContactText }}</span>
            </div>
          </div>

          <hr class="mt-4" />

          <div class="section__title-wrap">
            <h2 class="section__title section__title2">商品搜索</h2>
          </div>

          <div class="section__title-wrap search">
            <div class="input">
              <input
                v-model="searchInput"
                type="text"
                placeholder="输入关键词搜索商品"
                @keyup.enter="runSearch"
              />
            </div>

            <button type="button" class="search_button" @click="runSearch">
              <svg class="icon" viewBox="0 0 1024 1024" width="22" height="22" aria-hidden="true">
                <path
                  d="M455.253333 657.92c117.76 0 213.333333-95.573333 213.333334-213.333333s-95.573333-213.333333-213.333334-213.333334-213.333333 95.573333-213.333333 213.333334 95.573333 213.333333 213.333333 213.333333z m229.76-22.4l169.813334 169.813333c16.64 16.64 16.64 43.733333 0 60.373334-16.64 16.64-43.733333 16.64-60.373334 0l-172.8-172.8c-47.573333 32-104.746667 50.56-166.4 50.56-164.906667 0-298.666667-133.76-298.666666-298.666667s133.76-298.666667 298.666666-298.666667 298.666667 133.76 298.666667 298.666667c0 72.32-25.813333 138.88-68.906667 190.72z"
                  fill="#ffffff"
                />
              </svg>
              <span>商品查询</span>
            </button>
          </div>

          <div v-if="activeKeywords" class="search-chip-row">
            <span class="search-chip">当前搜索：{{ activeKeywords }}</span>
            <button type="button" class="search-clear" @click="clearSearch">清空搜索</button>
          </div>

          <div class="section__title-wrap">
            <h2 class="section__title">商品分类</h2>
          </div>

          <div class="category_list">
            <button
              v-for="category in categories"
              :key="category.id"
              type="button"
              class="category_box"
              :class="{ active: !activeKeywords && category.id === selectedCategoryId }"
              @click="selectCategory(category)"
            >
              <div class="card">
                <div class="card__title pb-0">
                  <h3>{{ category.name }}</h3>
                </div>
                <div class="card__content">
                  <span class="card__s_cateremark">共{{ category.goods_count }}种商品</span>
                </div>
                <span v-if="!activeKeywords && category.id === selectedCategoryId" class="corner-check"></span>
              </div>
            </button>
          </div>

          <hr class="mt-4" />

          <div class="section__title-wrap">
            <h2 class="section__title section__title--pre">{{ activeKeywords ? '搜索结果' : '选择商品' }}</h2>
          </div>

          <div v-if="goodsLoading" class="panel-empty">正在加载商品...</div>
          <div v-else-if="goodsList.length === 0" class="panel-empty">当前条件下暂无商品</div>
          <div v-else class="goods_list">
            <button
              v-for="goods in goodsList"
              :key="goods.goods_key"
              type="button"
              class="goods_box"
              :class="{ active: goods.goods_key === selectedProductKey }"
              @click="selectProduct(goods)"
            >
              <div class="card">
                <div class="card__title">
                  <h3>{{ goods.name }}</h3>
                </div>

                <div class="goods-card-desc">
                  {{ stripHtml(goods.description) }}
                </div>

                <div class="card__content goods-card-content">
                  <span class="card__detail_price">￥{{ formatAmount(goods.price) }}</span>
                  <span class="card__detail_stock">{{ resolveGoodsSubline(goods) }}</span>
                </div>

                <span v-if="goods.goods_key === selectedProductKey" class="corner-check"></span>
              </div>
            </button>
          </div>

          <div v-if="saleMessages.length" class="mt-4">
            <div class="d-flex align-items-center title-row">
              <span class="title-icon">✦</span>
              <h2 class="section__title_2 mb-0">商品优惠</h2>
            </div>
            <div class="text_box sale_message">
              <span v-for="message in saleMessages" :key="message">{{ message }}</span>
            </div>
          </div>

          <div class="mt-4">
            <div class="d-flex align-items-center title-row">
              <span class="title-icon title-icon-blue">◼</span>
              <h2 class="section__title_2 mb-0">商品描述</h2>
            </div>

            <div class="text_box mt-3" id="remark">
              <h3 class="product-title">{{ selectedProduct?.name || '请选择商品' }}</h3>
              <p>{{ selectedProductDescription }}</p>
            </div>
          </div>

          <hr class="mt-4" />

          <div class="row mt-4">
            <div class="col-12" id="order_box">
              <div class="ure_info_box">
                <div class="ure_info_hide">
                  <span>填写信息/购买方式</span>
                </div>

                <div class="ure_info">
                  <div class="ure_info_input_box">
                    <div class="ure_info_input_box_hide">
                      <p>*</p>
                      <p>联系方式</p>
                    </div>

                    <div class="input">
                      <input
                        v-model="contact"
                        type="text"
                        placeholder="[必填]请填写 Bridgemind 注册邮箱"
                      />
                    </div>

                    <div class="msg">联系方式特别重要，可用来自动识别充值账号与查询订单</div>
                  </div>

                  <div class="ure_info_input_box d-flex purchase-toggle-row">
                    <div v-if="canUseCoupon" class="ure_info_input_box_hide btn-type">
                      <div
                        :class="{ on: couponEnabled }"
                        @click="couponEnabled = !couponEnabled"
                      >
                        <label>使用优惠券</label>
                      </div>
                    </div>

                    <div class="ure_info_input_box_hide btn-type">
                      <div class="static-chip">
                        <label>{{ selectedProduct?.contact_format === 'email' ? '邮箱到账' : '自动发货' }}</label>
                      </div>
                    </div>
                  </div>

                  <div v-if="couponEnabled && canUseCoupon" class="ure_info_input_box">
                    <div class="input">
                      <input v-model="couponCode" type="text" placeholder="请填写你的优惠券" />
                    </div>
                  </div>
                </div>

                <div class="pay_type">
                  <div class="pay_type_hide">选择支付方式</div>

                  <div class="pay_type_box">
                    <button
                      v-for="channel in channels"
                      :key="channel.id"
                      type="button"
                      class="pay_type_leng"
                      :class="{ pay_type_leng_xz: channel.id === selectedChannelId }"
                      @click="selectedChannelId = channel.id"
                    >
                      <img
                        v-if="channel.paytype?.icon"
                        :src="channel.paytype.icon"
                        :alt="channel.show_name"
                      />
                      <span>{{ channel.show_name }}</span>
                      <span v-if="channel.id === selectedChannelId" class="pay-selected-mark"></span>
                    </button>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </template>
      </div>
    </section>

    <footer>
      <div class="container">
        <p>Copyright © 2024-2026 Bridgemind All rights reserved.</p>
      </div>
    </footer>

    <section class="section_buttom">
      <div class="container bottom-inner">
        <div class="goods_name">
          <svg class="icon" viewBox="0 0 1024 1024" width="28" height="28" aria-hidden="true">
            <path
              d="M847.872 734.208L784.384 368.64H239.616l-63.488 365.568c-9.216 51.2 30.72 105.472 81.92 105.472h506.88c52.224 0 92.16-54.272 82.944-105.472zM512 648.192c-71.68 0-130.048-58.368-130.048-130.048 0-6.144 4.096-10.24 10.24-10.24s10.24 4.096 10.24 10.24c0 60.416 49.152 109.568 109.568 109.568s109.568-49.152 109.568-109.568c0-6.144 4.096-10.24 10.24-10.24s10.24 4.096 10.24 10.24c0 71.68-58.368 130.048-130.048 130.048z"
              fill="#2F8AF5"
            />
            <path
              d="M791.552 358.4l-19.456-96.256c-7.168-39.936-46.08-67.584-87.04-67.584H340.992c-40.96 0-74.752 28.672-81.92 68.608L243.712 358.4h547.84z"
              fill="#83C1FF"
            />
          </svg>
          <span>{{ selectedProduct?.name || '请选择商品' }}</span>
        </div>

        <div class="shuliang_box d-flex align-items-center">
          <button type="button" class="btn" @click="decreaseQuantity">
            <span class="minus-bar"></span>
          </button>

          <div class="input">
            <input
              v-model.number="quantity"
              type="number"
              min="1"
              inputmode="numeric"
              @blur="normalizeQuantity"
              @keyup.enter="normalizeQuantity"
            />
          </div>

          <button type="button" class="btn" @click="increaseQuantity">
            <span></span>
            <span></span>
          </button>

          <div class="jiage d-flex align-items-center">
            <span>支付金额 :</span>
            <span>￥</span>
            <s v-if="showOriginalAmount">{{ formatAmount(originalAmount) }}</s>
            <span>{{ formatAmount(totalAmount) }}</span>
          </div>
        </div>

        <div class="queding_btn">
          <button
            type="button"
            name="check_pay"
            class="check_pay"
            :disabled="ordering || !selectedProduct"
            @click="submitOrder"
          >
            {{ ordering ? '正在创建订单' : '确认支付' }}
          </button>
        </div>
      </div>
    </section>

    <div v-if="agreementVisible" class="agreement-overlay">
      <div class="agreement-modal">
        <div class="agreement-header">
          <h3>购卡协议</h3>
          <button type="button" class="agreement-close" @click="closeAgreement">×</button>
        </div>

        <div class="agreement-body">
          <p>
            <strong>温馨提示：</strong>
            请填写 Bridgemind 注册邮箱，支付完成后系统将自动处理充值。
          </p>
          <p class="agreement-emphasis">
            如有异常，请先通过订单查询确认状态，再联系站点管理员处理。
          </p>
          <ol>
            <li>余额充值与并发升级均为虚拟服务，付款后不支持随意退款。</li>
            <li>联系方式建议填写 Bridgemind 注册邮箱，便于自动识别与售后排查。</li>
            <li>若支付页未自动打开，可再次点击底部“确认支付”重新拉起支付链接。</li>
          </ol>
        </div>

        <div class="agreement-footer">
          <button type="button" @click="acceptAgreement">我同意协议</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAppStore } from '@/stores'
import { useAuthStore } from '@/stores/auth'

interface ApiResponse<T> {
  code: number
  msg: string
  data: T
}

interface ShopCategory {
  id: number
  name: string
  image?: string
  goods_count: number
}

interface ShopInfo {
  create_time: number
  nickname: string
  avatar?: string
  description: string
  token: string
  auth_status: number
  contact_qq?: string
  contact_mobile?: string
  contact_wechat?: string
}

interface ShopUserInfo {
  link?: string
  nickname?: string
  avatar?: string
  token?: string
}

interface ShopGoodsCategory {
  id: number
  name: string
}

interface ShopGoodsExtend {
  stock_count?: number
  show_stock_type?: number
  send_order?: number
  limit_count?: number
  query_password_status?: number
}

interface ShopGoodsDiscount {
  available: number
  rebate: number
}

interface ShopGoods {
  link: string
  goods_key: string
  name: string
  price: number
  market_price: number
  description: string
  image?: string
  contact_format?: string
  coupon_status?: number
  category?: ShopGoodsCategory
  user?: ShopUserInfo
  discount?: ShopGoodsDiscount | null
  multipleoffers?: { available: number; discount_type: number } | null
  fullgift?: { available: number } | null
  extend?: ShopGoodsExtend | null
}

interface GoodsListPayload {
  total: number
  list: ShopGoods[]
}

interface ShopChannel {
  id: number
  show_name: string
  rate?: number
  paytype?: {
    icon?: string
  } | null
}

interface PriceInfo {
  original_amount: number
  total_amount: number
  fee: number
  fee_payer: number
  sales_style?: Record<string, string> | string[]
  coupon_available?: number
  coupon_price?: number
}

interface CreateOrderResponse {
  payurl?: string
  trade_no?: string
  total_amount: number
}

const DEFAULT_SHOP_URL = 'https://pay.ldxp.cn/shop/HQP8RZ4F'
const AGREEMENT_STORAGE_KEY = 'bridgemind.recharge.agreement.v3'

const route = useRoute()
const router = useRouter()
const appStore = useAppStore()
const authStore = useAuthStore()

const loading = ref(true)
const goodsLoading = ref(false)
const ordering = ref(false)
const agreementVisible = ref(false)

const merchantInfo = ref<ShopInfo | null>(null)
const categories = ref<ShopCategory[]>([])
const goodsList = ref<ShopGoods[]>([])
const channels = ref<ShopChannel[]>([])
const priceInfo = ref<PriceInfo | null>(null)

const selectedCategoryId = ref<number | null>(null)
const selectedProductKey = ref('')
const selectedChannelId = ref<number | null>(null)

const quantity = ref(1)
const contact = ref('')
const couponEnabled = ref(false)
const couponCode = ref('')
const searchInput = ref('')
const activeKeywords = ref('')

let latestPriceRequestId = 0

const embeddedMode = computed(() => {
  if (route.query.ui_mode === 'embedded') return true
  if (typeof window === 'undefined') return false
  return window.self !== window.top
})

const siteName = computed(() => appStore.siteName || appStore.cachedPublicSettings?.site_name || 'Bridgemind')
const docUrl = computed(() => appStore.docUrl || appStore.cachedPublicSettings?.doc_url || '')

const sourceUrl = computed(() => {
  const raw = typeof route.query.src_url === 'string' ? route.query.src_url : ''
  return sanitizeURL(raw, '')
})

const shopUrl = computed(() => {
  const raw = typeof route.query.shop_url === 'string' ? route.query.shop_url : DEFAULT_SHOP_URL
  return sanitizeURL(raw, DEFAULT_SHOP_URL)
})

const shopOrigin = computed(() => {
  try {
    return new URL(shopUrl.value).origin
  } catch {
    return 'https://pay.ldxp.cn'
  }
})

const shopToken = computed(() => {
  if (typeof route.query.shop_token === 'string' && route.query.shop_token.trim()) {
    return route.query.shop_token.trim()
  }
  return extractShopToken(shopUrl.value)
})

const orderQueryUrl = computed(() => buildShopUrl('/order'))

const merchantContactValue = computed(() => {
  const info = merchantInfo.value
  return (
    info?.contact_wechat?.trim() ||
    info?.contact_mobile?.trim() ||
    info?.contact_qq?.trim() ||
    appStore.contactInfo?.trim() ||
    ''
  )
})

const merchantContactLabel = computed(() => {
  const info = merchantInfo.value
  if (info?.contact_wechat?.trim()) return '商家微信'
  if (info?.contact_mobile?.trim()) return '商家手机'
  if (info?.contact_qq?.trim()) return '商家QQ'
  return '站点联系'
})

const merchantContactText = computed(() =>
  merchantContactValue.value ? `${merchantContactLabel.value}：${merchantContactValue.value}` : ''
)

const formattedCreateDate = computed(() =>
  merchantInfo.value?.create_time ? formatUnixDate(merchantInfo.value.create_time) : ''
)

const selectedProduct = computed(
  () => goodsList.value.find((goods) => goods.goods_key === selectedProductKey.value) || null
)

const selectedProductDescription = computed(() => {
  const description = stripHtml(selectedProduct.value?.description || '')
  if (description) return description
  return '请选择需要充值的商品，填写 Bridgemind 注册邮箱后即可发起支付。'
})

const minQuantity = computed(() => Math.max(1, selectedProduct.value?.extend?.limit_count || 1))
const canUseCoupon = computed(() => (selectedProduct.value?.coupon_status ?? 0) === 1)

const totalAmount = computed(() => {
  if (priceInfo.value) return priceInfo.value.total_amount
  if (!selectedProduct.value) return 0
  return selectedProduct.value.price * quantity.value
})

const originalAmount = computed(() => {
  if (priceInfo.value) return priceInfo.value.original_amount
  return totalAmount.value
})

const showOriginalAmount = computed(() => originalAmount.value > totalAmount.value)

const saleMessages = computed(() => {
  const salesStyle = priceInfo.value?.sales_style
  if (!salesStyle) return []
  if (Array.isArray(salesStyle)) {
    return salesStyle.map((item) => String(item))
  }
  return Object.entries(salesStyle).map(([key, value]) => `${key} ${value}`.trim())
})

const floatingActions = computed(() => {
  const actions: Array<{ key: string; label: string; href: string; tone: string }> = []
  if (orderQueryUrl.value) {
    actions.push({ key: 'order', label: '订单查询', href: orderQueryUrl.value, tone: 'danger' })
  }
  if (docUrl.value) {
    actions.push({ key: 'docs', label: '使用文档', href: docUrl.value, tone: 'support' })
  }
  if (shopUrl.value) {
    actions.push({ key: 'shop', label: '原始店铺', href: shopUrl.value, tone: 'group' })
  }
  return actions
})

function sanitizeURL(raw: string, fallback: string): string {
  if (!raw) return fallback
  try {
    const url = new URL(raw, window.location.origin)
    if (url.protocol !== 'http:' && url.protocol !== 'https:') {
      return fallback
    }
    return url.toString()
  } catch {
    return fallback
  }
}

function extractShopToken(url: string): string {
  try {
    const parsed = new URL(url)
    const parts = parsed.pathname.split('/').filter(Boolean)
    const shopIndex = parts.findIndex((part) => part === 'shop')
    if (shopIndex >= 0 && parts[shopIndex + 1]) {
      return parts[shopIndex + 1]
    }
  } catch {
    return ''
  }
  return ''
}

function buildShopUrl(path: string): string {
  try {
    return new URL(path, shopOrigin.value).toString()
  } catch {
    return shopUrl.value
  }
}

function stripHtml(value: string): string {
  return value
    .replace(/<[^>]+>/g, ' ')
    .replace(/&nbsp;/g, ' ')
    .replace(/\s+/g, ' ')
    .trim()
}

function formatUnixDate(unixSeconds: number): string {
  const date = new Date(unixSeconds * 1000)
  if (Number.isNaN(date.getTime())) return ''
  const year = date.getFullYear()
  const month = `${date.getMonth() + 1}`.padStart(2, '0')
  const day = `${date.getDate()}`.padStart(2, '0')
  return `${year}-${month}-${day}`
}

function formatAmount(value: number): string {
  return Number(value || 0).toFixed(2)
}

function resolveGoodsSubline(goods: ShopGoods): string {
  if ((goods.contact_format || '').toLowerCase() === 'email') {
    return '填写注册邮箱'
  }

  const stockCount = goods.extend?.stock_count ?? 0
  if (stockCount > 0) {
    return `库存${stockCount}份`
  }

  return '自动发货'
}

function navigateInternal(path: string) {
  if (typeof window !== 'undefined' && embeddedMode.value && window.top && window.top !== window) {
    window.top.location.href = new URL(path, window.location.origin).toString()
    return
  }
  router.push(path)
}

function openExternal(url: string, preferNewTab: boolean = false) {
  if (!url || typeof window === 'undefined') return

  if (preferNewTab && !embeddedMode.value) {
    const opened = window.open(url, '_blank', 'noopener,noreferrer')
    if (opened) {
      opened.opener = null
      return
    }
  }

  if (embeddedMode.value && window.top && window.top !== window) {
    window.top.location.href = url
    return
  }

  window.location.href = url
}

function navigateBack() {
  if (sourceUrl.value) {
    openExternal(sourceUrl.value, false)
    return
  }
  navigateInternal('/dashboard')
}

async function postShopApi<T>(path: string, payload: Record<string, unknown>): Promise<T> {
  const response = await fetch(buildShopUrl(path), {
    method: 'POST',
    mode: 'cors',
    credentials: 'omit',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(payload),
  })

  if (!response.ok) {
    throw new Error(`请求失败（${response.status}）`)
  }

  const result = (await response.json()) as ApiResponse<T>
  if (result.code !== 1) {
    throw new Error(result.msg || '请求失败')
  }

  return result.data
}

async function loadMerchantInfo() {
  merchantInfo.value = await postShopApi<ShopInfo>('/shopApi/Shop/info', {
    token: shopToken.value,
  })
}

async function loadChannels() {
  channels.value = await postShopApi<ShopChannel[]>('/shopApi/Shop/getUserChannel', {
    token: shopToken.value,
  })
  selectedChannelId.value = channels.value[0]?.id ?? null
}

async function loadCategories() {
  categories.value = await postShopApi<ShopCategory[]>('/shopApi/Shop/categoryList', {
    token: shopToken.value,
    goods_type: 'card',
  })

  if (!categories.value.length) {
    selectedCategoryId.value = null
    goodsList.value = []
    selectedProductKey.value = ''
    return
  }

  if (!selectedCategoryId.value || !categories.value.some((category) => category.id === selectedCategoryId.value)) {
    selectedCategoryId.value = categories.value[0].id
  }

  await loadGoods()
}

async function loadGoods() {
  goodsLoading.value = true

  try {
    const payload: Record<string, unknown> = {
      token: shopToken.value,
      goods_type: 'card',
      current: 1,
      pageSize: 999,
      keywords: activeKeywords.value,
    }

    if (!activeKeywords.value && selectedCategoryId.value !== null) {
      payload.category_id = selectedCategoryId.value
    }

    const goodsPayload = await postShopApi<GoodsListPayload>('/shopApi/Shop/goodsList', payload)
    goodsList.value = goodsPayload.list || []

    if (!goodsList.value.some((goods) => goods.goods_key === selectedProductKey.value)) {
      selectedProductKey.value = goodsList.value[0]?.goods_key ?? ''
    }
  } finally {
    goodsLoading.value = false
  }
}

async function loadPriceInfo() {
  const product = selectedProduct.value
  if (!product || !selectedChannelId.value) {
    priceInfo.value = null
    return
  }

  const requestId = ++latestPriceRequestId

  try {
    const data = await postShopApi<PriceInfo>('/shopApi/Shop/getGoodsPrice', {
      goods_key: product.goods_key,
      quantity: quantity.value,
      coupon_code: couponEnabled.value ? couponCode.value.trim() : '',
      channel_id: selectedChannelId.value,
    })

    if (requestId !== latestPriceRequestId) return
    priceInfo.value = data
  } catch (error) {
    if (requestId !== latestPriceRequestId) return
    priceInfo.value = null
    if (couponEnabled.value && couponCode.value.trim()) {
      appStore.showError((error as Error).message || '优惠券不可用')
    }
  }
}

function normalizeQuantity() {
  const parsed = Number.parseInt(String(quantity.value), 10)
  quantity.value = Number.isFinite(parsed) ? Math.max(minQuantity.value, parsed) : minQuantity.value
}

function increaseQuantity() {
  quantity.value += 1
}

function decreaseQuantity() {
  quantity.value = Math.max(minQuantity.value, quantity.value - 1)
}

function selectCategory(category: ShopCategory) {
  activeKeywords.value = ''
  searchInput.value = ''
  selectedCategoryId.value = category.id
  void loadGoods().catch((error: unknown) => {
    appStore.showError((error as Error).message || '加载商品失败')
  })
}

function selectProduct(goods: ShopGoods) {
  selectedProductKey.value = goods.goods_key
}

function runSearch() {
  activeKeywords.value = searchInput.value.trim()
  void loadGoods().catch((error: unknown) => {
    appStore.showError((error as Error).message || '加载商品失败')
  })
}

function clearSearch() {
  searchInput.value = ''
  activeKeywords.value = ''
  void loadGoods().catch((error: unknown) => {
    appStore.showError((error as Error).message || '加载商品失败')
  })
}

function closeAgreement() {
  agreementVisible.value = false
}

function acceptAgreement() {
  agreementVisible.value = false
  if (typeof window !== 'undefined') {
    window.sessionStorage.setItem(AGREEMENT_STORAGE_KEY, '1')
  }
}

async function copyMerchantContact() {
  if (!merchantContactValue.value) return

  try {
    await navigator.clipboard.writeText(merchantContactValue.value)
    appStore.showSuccess('联系方式已复制')
  } catch {
    appStore.showInfo(merchantContactText.value)
  }
}

async function submitOrder() {
  if (!selectedProduct.value) {
    appStore.showError('请先选择商品')
    return
  }

  const trimmedContact = contact.value.trim()
  if (!trimmedContact) {
    appStore.showError('请填写 Bridgemind 注册邮箱')
    return
  }

  if (
    (selectedProduct.value.contact_format || '').toLowerCase() === 'email' &&
    !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(trimmedContact)
  ) {
    appStore.showError('请输入有效的邮箱地址')
    return
  }

  if (!selectedChannelId.value) {
    appStore.showError('请选择支付方式')
    return
  }

  ordering.value = true

  try {
    const data = await postShopApi<CreateOrderResponse>('/shopApi/Pay/order', {
      goods_key: selectedProduct.value.goods_key,
      quantity: quantity.value,
      coupon_code: couponEnabled.value ? couponCode.value.trim() : '',
      channel_id: selectedChannelId.value,
      contact: trimmedContact,
    })

    if (data.total_amount === 0 && data.trade_no) {
      appStore.showSuccess('订单已创建')
      openExternal(buildShopUrl(`/order/result/${data.trade_no}`), false)
      return
    }

    if (!data.payurl) {
      throw new Error('支付链接生成失败')
    }

    appStore.showSuccess('订单已创建，正在打开支付页面')
    openExternal(data.payurl, true)
  } catch (error) {
    appStore.showError((error as Error).message || '创建订单失败')
  } finally {
    ordering.value = false
  }
}

async function initializeStore() {
  if (!shopToken.value) {
    appStore.showError('未找到店铺配置')
    loading.value = false
    return
  }

  loading.value = true

  try {
    await Promise.all([loadMerchantInfo(), loadChannels()])
    await loadCategories()
    await loadPriceInfo()
  } catch (error) {
    appStore.showError((error as Error).message || '同步店铺失败')
  } finally {
    loading.value = false
  }
}

watch(
  () => authStore.user?.email,
  (value) => {
    if (!contact.value && value) {
      contact.value = value
    }
  },
  { immediate: true },
)

watch(selectedProduct, (product) => {
  if (!product) {
    priceInfo.value = null
    return
  }

  if (quantity.value < minQuantity.value) {
    quantity.value = minQuantity.value
  }

  if (!canUseCoupon.value) {
    couponEnabled.value = false
    couponCode.value = ''
  }
})

watch([selectedProductKey, quantity, selectedChannelId, couponEnabled, couponCode], () => {
  void loadPriceInfo()
})

onMounted(async () => {
  if (!appStore.publicSettingsLoaded) {
    try {
      await appStore.fetchPublicSettings()
    } catch {
      // Keep the page usable even if public settings fail to refresh.
    }
  }

  if (typeof window !== 'undefined') {
    agreementVisible.value = window.sessionStorage.getItem(AGREEMENT_STORAGE_KEY) !== '1'
  }

  if (!contact.value) {
    contact.value = authStore.user?.email ?? ''
  }

  await initializeStore()
})
</script>

<style scoped>
.store-page {
  min-height: 100vh;
  background:
    linear-gradient(180deg, rgba(227, 236, 255, 0.92) 0%, rgba(243, 247, 255, 0.95) 26%, #eef3fb 100%);
  color: #545454;
  padding-bottom: 140px;
}

.store-page-embedded {
  padding-top: 0;
}

.container {
  width: min(1120px, calc(100% - 32px));
  margin: 0 auto;
}

.header {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  background-color: #fff;
  z-index: 99;
  transition: 0.5s;
  box-shadow: 0 1px 15px 0 rgba(0, 0, 0, 0.12);
}

.header-inner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
  min-height: 70px;
}

.header_left,
.header_right {
  display: flex;
  align-items: center;
  min-height: 70px;
}

.header_logo {
  display: inline-flex;
  align-items: center;
  gap: 18px;
  margin-right: 30px;
  height: auto;
  font-weight: 700;
  color: #535761;
  font-size: 18px;
  border: none;
  background: transparent;
  padding: 0;
  cursor: pointer;
}

.header_logo img {
  width: 48px;
  height: 48px;
}

.online-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  min-width: 85px;
  padding: 0 10px;
  height: 28px;
  background: #698df3;
  border: 2px solid #4274ff;
  box-shadow: 1px 4px 5px 0 hsla(0, 0%, 66.3%, 0.2);
  border-radius: 14px;
  font-size: 12px;
  color: #fff;
  text-align: center;
  cursor: pointer;
}

.header_button {
  display: flex;
  flex-direction: row;
  justify-content: center;
  align-items: center;
  width: 130px;
  height: 44px;
  border: none;
  border-radius: 50px;
  background: linear-gradient(0deg, #2a62ff, #4e7dff);
  box-shadow: 0 5px 6px 0 rgba(73, 105, 230, 0.22);
}

.header_button span {
  display: block;
  font-size: 14px;
  color: #fff;
  letter-spacing: 0.4px;
  margin-left: 5px;
}

footer {
  margin-bottom: 85px;
}

footer p {
  text-align: center;
  font-size: 14px;
  color: #585963;
}

.section {
  position: relative;
  padding-top: 60px;
}

.section--first {
  padding-top: 110px;
}

.section--last {
  padding-bottom: 50px;
}

.section .container {
  box-shadow: 0 7px 29px 0 rgba(18, 52, 91, 0.11);
  border-radius: 10px;
  padding: 20px 35px 32px;
  background: #fff;
}

.section__title-wrap {
  position: relative;
  display: flex;
  flex-direction: column;
  justify-content: flex-start;
  align-items: flex-start;
  margin-top: 20px;
}

.section__title {
  color: #545454;
  font-weight: 700;
  font-size: 18px;
  line-height: 100%;
  margin-bottom: 0;
  position: relative;
  padding-left: 25px;
  font-family: 'Montserrat', sans-serif;
}

.section__title2,
.section .section__title_2 {
  font-size: 14px;
  color: #545454;
}

.section__title::before {
  content: '';
  position: absolute;
  display: block;
  top: 2px;
  bottom: 2px;
  left: 0;
  width: 3px;
  background-color: #3369ff;
  border-radius: 4px;
}

.section__title--pre::before {
  background-color: #f26c2a;
}

.section hr {
  border: none;
  height: 1px;
  background-color: #f7f7f7;
}

.merchant-meta-row {
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
  margin-top: 18px;
}

.merchant-meta-item {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 14px;
  font-weight: 600;
  color: #545454;
}

.merchant-meta-item strong {
  color: #3369ff;
}

.merchant-meta-icon {
  color: #3369ff;
  font-size: 13px;
}

.search {
  display: flex;
  flex-direction: row;
  align-items: center;
  gap: 16px;
}

.search .input,
#order_box .ure_info_box .ure_info .ure_info_input_box .input {
  background: #fff;
  border: 1px solid #f0f0f0;
  box-shadow: 0 4px 10px 0 rgba(135, 142, 154, 0.07);
  border-radius: 4px;
  overflow: hidden;
}

.search .input {
  height: 40px;
  width: 400px;
}

.search .input input,
#order_box .ure_info_box .ure_info .ure_info_input_box .input input {
  display: inline-block;
  width: 100%;
  padding: 0 20px;
  height: 100%;
  font-weight: 500;
  border: none;
  outline: none;
}

.search_button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  width: 110px;
  height: 32px;
  border: none;
  border-radius: 16px;
  background: linear-gradient(0deg, #2a62ff, #4e7dff);
  box-shadow: 0 5px 6px 0 rgba(73, 105, 230, 0.22);
  color: #ffffff;
  cursor: pointer;
  font-size: 14px;
}

.search-chip-row {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-top: 16px;
}

.search-chip {
  display: inline-flex;
  align-items: center;
  height: 32px;
  padding: 0 14px;
  border-radius: 16px;
  background: #edf3ff;
  color: #3369ff;
  font-size: 13px;
  font-weight: 600;
}

.search-clear {
  border: none;
  background: transparent;
  color: #999;
  font-size: 13px;
  cursor: pointer;
}

.category_list {
  display: flex;
  gap: 16px;
  max-height: 420px;
  overflow-x: auto;
  padding: 0 0 6px;
}

.category_box,
.goods_box {
  padding: 0;
  border: none;
  background: transparent;
  text-align: left;
  cursor: pointer;
}

.category_box {
  min-width: 220px;
}

.card {
  position: relative;
  display: block;
  margin-top: 20px;
  border-radius: 10px;
  overflow: hidden;
  background-color: #fff;
  transition: 0.3s;
  border: 2px solid #f1f4fb;
  box-shadow: 0 4px 10px 0 rgba(135, 142, 154, 0.14);
}

.card:hover {
  box-shadow: 0 1px 15px 0 rgba(0, 0, 0, 0.2);
}

.card__title {
  position: relative;
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: flex-start;
}

.card__title h3 {
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
  width: 100%;
  margin-bottom: 10px;
  color: #33373f;
  font-size: 15px;
  font-weight: 600;
}

.card__content {
  display: flex;
  flex-direction: row;
  justify-content: space-between;
  align-items: center;
  width: 100%;
}

.card__s_cateremark {
  color: #999;
  font-size: 15px;
}

.category_box .card {
  min-height: 124px;
  padding: 18px 18px 20px;
}

.category_box.active .card {
  background: linear-gradient(-45deg, #3369ff, #3798f7);
  box-shadow: 0 7px 10px 0 rgba(54, 144, 248, 0.23);
  border: 1px solid #3580fb;
}

.category_box.active .card .card__title h3,
.category_box.active .card .card__content .card__s_cateremark {
  color: #fff;
}

.goods_list {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 16px;
}

.goods_box .card {
  min-height: 172px;
  padding: 18px;
}

.goods_box.active .card {
  border: 2px solid #3369ff;
}

.goods-card-desc {
  min-height: 44px;
  font-size: 13px;
  line-height: 1.6;
  color: #7d8597;
  margin: 4px 0 12px;
}

.goods-card-content {
  align-items: flex-end;
}

.card__detail_price {
  font-weight: 700;
  font-size: 26px;
  color: #3369ff;
  line-height: 1;
}

.card__detail_stock {
  color: #0db26a;
  font-size: 13px;
  font-weight: 600;
}

.corner-check {
  position: absolute;
  right: 0;
  bottom: 0;
  width: 0;
  height: 0;
  border-left: 26px solid transparent;
  border-bottom: 26px solid #3369ff;
}

.corner-check::after {
  content: '';
  position: absolute;
  right: 2px;
  bottom: -22px;
  width: 8px;
  height: 12px;
  border-right: 2px solid #fff;
  border-bottom: 2px solid #fff;
  transform: rotate(40deg);
}

.title-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.title-icon {
  color: #f2a52a;
  font-size: 18px;
}

.title-icon-blue {
  color: #2b8cff;
}

.text_box {
  padding: 18px 20px;
  background: #fbfcff;
  border: 1px solid #eef2fb;
  border-radius: 12px;
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.75);
}

.sale_message span {
  display: block;
  font-size: 14px;
  color: #3369ff;
}

.sale_message span + span {
  margin-top: 8px;
}

.product-title {
  margin: 0 0 12px;
  font-size: 18px;
  font-weight: 700;
  color: #151f34;
}

.text_box p {
  margin: 0;
  font-size: 15px;
  line-height: 1.8;
  color: #6a7386;
}

#order_box .ure_info_box .ure_info_hide span {
  margin-left: 3px;
  color: #545454;
  font-size: 16px;
  font-weight: 700;
}

#order_box .ure_info_box .ure_info {
  margin: 0 auto;
  padding-bottom: 25px;
  border-bottom: 1px solid #f7f7f7;
  padding-top: 27px;
}

#order_box .ure_info_box .ure_info .ure_info_input_box {
  margin-left: 7px;
  margin-top: 20px;
}

#order_box .ure_info_box .ure_info .ure_info_input_box:first-child {
  margin-top: 0;
}

#order_box .ure_info_box .ure_info .ure_info_input_box .ure_info_input_box_hide {
  font-weight: 700;
}

#order_box .ure_info_box .ure_info .ure_info_input_box .ure_info_input_box_hide p {
  display: inline-block;
  vertical-align: middle;
}

#order_box .ure_info_box .ure_info .ure_info_input_box .ure_info_input_box_hide p:first-child {
  font-size: 12px;
  color: #fb636b;
  width: 20px;
}

#order_box .ure_info_box .ure_info .ure_info_input_box .ure_info_input_box_hide p:nth-child(2) {
  color: #545454;
  font-size: 14px;
}

#order_box .ure_info_box .ure_info .ure_info_input_box .input {
  height: 50px;
  width: 430px;
  margin-top: 11px;
}

#order_box .ure_info_box .ure_info .ure_info_input_box .msg {
  margin-top: 14px;
  font-size: 12px;
  font-weight: 500;
  color: #999;
  margin-left: 6px;
}

.purchase-toggle-row {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.btn-type {
  display: flex;
  align-items: center;
}

.btn-type > div {
  margin-right: 10px;
  border-radius: 45px;
  background: #f8f8f8;
  border: 2px solid #f6f6f6;
  color: #999;
  cursor: pointer;
}

.btn-type > div.on {
  border-color: #3369ff;
  color: #3369ff;
}

.btn-type label {
  display: inline-block;
  padding: 8px 16px;
  cursor: pointer;
  margin-bottom: 0;
  font-size: 14px;
}

.static-chip {
  color: #3369ff;
  border-color: #d9e6ff;
  background: #f3f8ff;
}

#order_box .ure_info_box .pay_type .pay_type_hide {
  margin-left: 22px;
  color: #999;
  font-size: 14px;
  font-weight: 700;
  padding-top: 22px;
}

#order_box .ure_info_box .pay_type .pay_type_box {
  margin-top: 21px;
  margin-left: 17px;
  display: flex;
  flex-wrap: wrap;
  gap: 11px;
}

#order_box .ure_info_box .pay_type .pay_type_box .pay_type_leng {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  line-height: 43px;
  width: 148px;
  text-align: center;
  background: #f8f8f8;
  border: 2px solid #f6f6f6;
  border-radius: 4px;
  position: relative;
  margin-bottom: 8px;
}

#order_box .ure_info_box .pay_type .pay_type_box .pay_type_leng img {
  width: 21px;
  height: 21px;
}

#order_box .ure_info_box .pay_type .pay_type_box .pay_type_leng span {
  font-weight: 700;
  font-size: 13px;
  color: #545454;
}

#order_box .ure_info_box .pay_type .pay_type_box .pay_type_leng_xz {
  border-color: #3369ff;
}

.pay-selected-mark {
  position: absolute;
  right: -2px;
  bottom: -2px;
  width: 16px;
  height: 16px;
  background: linear-gradient(135deg, #5c91ff, #2b67ff);
  clip-path: polygon(100% 0, 0 100%, 100% 100%);
}

.section_buttom {
  width: 100%;
  min-height: 62px;
  background: #fff;
  box-shadow: 0 -1px 0 0 hsla(0, 0%, 71%, 0.18);
  position: fixed;
  left: 0;
  bottom: 0;
  z-index: 80;
}

.bottom-inner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 18px;
  min-height: 62px;
}

.goods_name {
  display: inline-flex;
  align-items: center;
  min-width: 220px;
  overflow: hidden;
}

.goods_name span {
  color: #545454;
  font-weight: 700;
  font-size: 14px;
  margin-left: 20px;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
}

.shuliang_box {
  display: inline-flex;
  align-items: center;
}

.shuliang_box .btn {
  width: 30px;
  height: 30px;
  background: #f8f8f8;
  border: 1px solid #f6f6f6;
  border-radius: 50%;
  position: relative;
  cursor: pointer;
  user-select: none;
}

.shuliang_box .input {
  margin-left: 20px;
  margin-right: 20px;
  background: #f8f8f8;
  border: 1px solid #f0f0f0;
  border-radius: 4px;
  height: 37px;
  width: 80px;
}

.shuliang_box .input input {
  background: #f8f8f8;
  border: none;
  width: 100%;
  color: #545454;
  font-weight: 700;
  font-size: 18px;
  height: 100%;
  text-align: center;
  outline: none;
}

.shuliang_box .input input::-webkit-outer-spin-button,
.shuliang_box .input input::-webkit-inner-spin-button {
  -webkit-appearance: none;
  margin: 0;
}

.shuliang_box .btn span:first-child {
  width: 14px;
  height: 2px;
  margin-left: -7px;
  margin-top: -1px;
}

.shuliang_box .btn span:nth-child(2) {
  width: 2px;
  height: 14px;
  margin-left: -1px;
  margin-top: -7px;
}

.shuliang_box .btn span {
  background: #545454;
  position: absolute;
  left: 50%;
  top: 50%;
}

.minus-bar {
  display: block;
}

.jiage {
  margin-left: 50px;
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.jiage span:first-child {
  font-size: 12px;
  color: #545454;
  font-weight: 700;
}

.jiage span:nth-child(2),
.jiage > span:last-child {
  font-size: 22px;
  color: #3369ff;
  font-weight: 700;
}

.jiage s {
  font-size: 12px;
  color: #b7b7b7;
  margin-left: 12px;
}

.queding_btn {
  display: flex;
  justify-content: center;
  align-items: center;
  width: 130px;
  height: 40px;
  border-radius: 50px;
  background: #3369ff;
  box-shadow: 0 5px 6px 0 rgba(73, 105, 230, 0.22);
  color: #fff;
}

.check_pay {
  width: 100%;
  height: 100%;
  border: none;
  background: transparent;
  color: #fff;
  font-size: 14px;
  font-weight: 700;
  cursor: pointer;
}

.check_pay:disabled {
  cursor: not-allowed;
  opacity: 0.72;
}

.agreement-overlay {
  position: fixed;
  inset: 0;
  background: rgba(15, 23, 42, 0.48);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 120;
  padding: 20px;
}

.agreement-modal {
  width: min(360px, 100%);
  max-height: min(520px, 85vh);
  background: #fff;
  border-radius: 10px;
  overflow: hidden;
  box-shadow: 0 24px 80px rgba(15, 23, 42, 0.24);
  display: flex;
  flex-direction: column;
}

.agreement-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 18px 18px 14px;
  border-bottom: 1px solid #f1f5f9;
}

.agreement-header h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 700;
  color: #4c5b78;
}

.agreement-close {
  border: none;
  background: transparent;
  font-size: 28px;
  line-height: 1;
  color: #475569;
  cursor: pointer;
}

.agreement-body {
  padding: 18px;
  overflow: auto;
  font-size: 15px;
  line-height: 1.8;
  color: #334155;
}

.agreement-body p {
  margin: 0 0 14px;
}

.agreement-emphasis {
  color: #ef4444;
  font-weight: 700;
}

.agreement-body ol {
  margin: 0;
  padding-left: 22px;
}

.agreement-body li + li {
  margin-top: 10px;
}

.agreement-footer {
  padding: 14px 18px 18px;
  display: flex;
  justify-content: flex-end;
}

.agreement-footer button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 110px;
  height: 36px;
  border: none;
  border-radius: 6px;
  background: linear-gradient(180deg, #4e7dff, #2a62ff);
  color: #fff;
  font-size: 14px;
  font-weight: 700;
  cursor: pointer;
}

.store-loading,
.panel-empty {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 180px;
  color: #64748b;
  font-size: 15px;
  font-weight: 600;
}

.loading-dot {
  width: 14px;
  height: 14px;
  border-radius: 999px;
  background: #3369ff;
  margin-right: 10px;
  box-shadow: 0 0 0 8px rgba(51, 105, 255, 0.12);
  animation: pulse 1.2s ease-in-out infinite;
}

.bangz_box {
  position: fixed;
  left: 0;
  top: 100px;
  z-index: 45;
}

.bangz_box .item {
  display: block;
  padding: 10px 14px;
  border-radius: 0 7px 7px 0;
  margin-top: 18px;
  color: #fff;
  text-decoration: none;
  box-shadow: 0 7px 10px 0 rgba(54, 144, 248, 0.23);
}

.bangz_box .item-danger {
  background: linear-gradient(45deg, #fd0b27, #ff4a4a);
}

.bangz_box .item-support {
  background: linear-gradient(45deg, #f737e8, #3369ff);
}

.bangz_box .item-group {
  background: linear-gradient(45deg, #3798f7, #3369ff);
}

.bangz_box span {
  font-weight: 600;
  font-size: 14px;
}

@keyframes pulse {
  0%,
  100% {
    transform: scale(1);
  }
  50% {
    transform: scale(1.08);
  }
}

@media (max-width: 1100px) {
  .goods_list {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }

  .bangz_box {
    display: none;
  }
}

@media (max-width: 900px) {
  .header-inner {
    flex-wrap: wrap;
    justify-content: center;
    padding: 10px 0;
  }

  .header_left,
  .header_right {
    min-height: auto;
  }

  .section--first {
    padding-top: 145px;
  }

  .section .container {
    padding: 18px 18px 28px;
  }

  .search {
    flex-direction: column;
    align-items: stretch;
  }

  .search .input,
  #order_box .ure_info_box .ure_info .ure_info_input_box .input {
    width: 100%;
  }

  .search_button {
    width: 100%;
    height: 40px;
    border-radius: 12px;
  }

  .goods_list {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .bottom-inner {
    flex-wrap: wrap;
    padding: 12px 0;
  }

  .goods_name {
    min-width: 100%;
  }

  .shuliang_box {
    width: 100%;
    justify-content: space-between;
  }

  .jiage {
    margin-left: 16px;
  }

  .queding_btn {
    width: 100%;
  }
}

@media (max-width: 640px) {
  .container {
    width: min(100%, calc(100% - 20px));
  }

  .header_logo {
    margin-right: 0;
    font-size: 16px;
  }

  .header_logo img {
    width: 40px;
    height: 40px;
  }

  .online-btn {
    width: 100%;
    justify-content: center;
    margin-top: 8px;
  }

  .section--first {
    padding-top: 170px;
  }

  .category_box {
    min-width: 180px;
  }

  .goods_list {
    grid-template-columns: 1fr;
  }

  .merchant-meta-row {
    gap: 10px;
  }

  .merchant-meta-item {
    width: 100%;
  }

  .shuliang_box {
    flex-wrap: wrap;
    row-gap: 12px;
  }

  .shuliang_box .input {
    margin: 0 12px;
  }

  .jiage {
    width: 100%;
    margin-left: 0;
    justify-content: flex-start;
  }
}
</style>
