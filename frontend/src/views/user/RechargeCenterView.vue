<template>
  <div
    class="recharge-center-shell"
    :class="embeddedMode ? 'recharge-center-embedded' : 'recharge-center-standalone'"
  >
    <div class="recharge-center-orb recharge-center-orb-left" aria-hidden="true"></div>
    <div class="recharge-center-orb recharge-center-orb-right" aria-hidden="true"></div>

    <div class="relative mx-auto flex min-h-full w-full max-w-6xl flex-col gap-6">
      <div class="flex flex-wrap items-center justify-between gap-3">
        <button
          type="button"
          class="recharge-inline-link"
          @click="navigateBack"
        >
          <svg class="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M15 18l-6-6 6-6" />
          </svg>
          <span>{{ t('rechargeCenter.back') }}</span>
        </button>

        <div class="flex flex-wrap items-center gap-2">
          <span class="recharge-chip">{{ siteName }}</span>
          <span class="recharge-chip recharge-chip-muted">{{ userLabel }}</span>
        </div>
      </div>

      <section class="recharge-hero">
        <div class="grid gap-6 lg:grid-cols-[1.15fr_0.85fr]">
          <div class="space-y-5">
            <div class="space-y-3">
              <p class="recharge-kicker">Recharge Hub</p>
              <h1 class="text-3xl font-black tracking-tight text-slate-950 sm:text-5xl dark:text-white">
                {{ t('rechargeCenter.title') }}
              </h1>
              <p class="max-w-2xl text-sm leading-7 text-slate-600 sm:text-base dark:text-slate-300">
                {{ t('rechargeCenter.subtitle') }}
              </p>
            </div>

            <div class="flex flex-wrap gap-2">
              <span class="recharge-pill">
                <span class="recharge-pill-label">{{ t('rechargeCenter.balanceLabel') }}</span>
                <strong>{{ balanceLabel }}</strong>
              </span>
              <span class="recharge-pill">
                <span class="recharge-pill-label">{{ t('rechargeCenter.statusLabel') }}</span>
                <strong>{{ paymentStatusLabel }}</strong>
              </span>
              <span v-if="contactLabel" class="recharge-pill">
                <span class="recharge-pill-label">{{ t('rechargeCenter.contactLabel') }}</span>
                <strong>{{ contactLabel }}</strong>
              </span>
            </div>

            <p class="text-sm leading-6 text-slate-500 dark:text-slate-400">
              {{ helperText }}
            </p>
          </div>

          <div class="recharge-summary">
            <p class="recharge-summary-kicker">{{ t('rechargeCenter.summaryTitle') }}</p>
            <div class="space-y-3">
              <div
                v-for="item in summaryItems"
                :key="item.label"
                class="recharge-summary-row"
              >
                <span class="recharge-summary-index">{{ item.index }}</span>
                <div>
                  <p class="text-sm font-semibold text-slate-900 dark:text-white">{{ item.label }}</p>
                  <p class="mt-1 text-xs leading-5 text-slate-500 dark:text-slate-400">{{ item.desc }}</p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      <section class="grid gap-4 lg:grid-cols-2">
        <a
          :href="shopUrl"
          target="_blank"
          rel="noopener noreferrer"
          class="recharge-card recharge-card-primary"
        >
          <span class="recharge-card-badge">{{ t('rechargeCenter.shop.badge') }}</span>
          <div class="recharge-card-icon">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7">
              <path stroke-linecap="round" stroke-linejoin="round" d="M12 3l2.6 5.4L20 9l-4 3.9.95 5.6L12 15.8 7.05 18.5 8 12.9 4 9l5.4-.6L12 3z" />
            </svg>
          </div>
          <div class="space-y-2">
            <h2 class="text-2xl font-black text-slate-950 dark:text-white">
              {{ t('rechargeCenter.shop.title') }}
            </h2>
            <p class="text-sm leading-6 text-slate-600 dark:text-slate-300">
              {{ t('rechargeCenter.shop.description') }}
            </p>
          </div>
          <div class="recharge-card-footer">
            <span>{{ t('rechargeCenter.shop.cta') }}</span>
            <svg class="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M5 12h14M13 5l7 7-7 7" />
            </svg>
          </div>
        </a>

        <button
          type="button"
          class="recharge-card recharge-card-secondary text-left"
          @click="openSecondaryChannel"
        >
          <span class="recharge-card-badge recharge-card-badge-soft">{{ secondaryBadge }}</span>
          <div class="recharge-card-icon recharge-card-icon-secondary">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7">
              <path stroke-linecap="round" stroke-linejoin="round" d="M4 7.5A2.5 2.5 0 016.5 5h11A2.5 2.5 0 0120 7.5v9A2.5 2.5 0 0117.5 19h-11A2.5 2.5 0 014 16.5v-9z" />
              <path stroke-linecap="round" stroke-linejoin="round" d="M8 9.5h8M8 13h5" />
            </svg>
          </div>
          <div class="space-y-2">
            <h2 class="text-2xl font-black text-slate-950 dark:text-white">
              {{ secondaryTitle }}
            </h2>
            <p class="text-sm leading-6 text-slate-600 dark:text-slate-300">
              {{ secondaryDescription }}
            </p>
          </div>
          <div class="recharge-card-footer">
            <span>{{ secondaryCta }}</span>
            <svg class="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M5 12h14M13 5l7 7-7 7" />
            </svg>
          </div>
        </button>
      </section>

      <section class="grid gap-4 lg:grid-cols-[0.95fr_1.05fr]">
        <div class="recharge-panel">
          <p class="recharge-panel-title">{{ t('rechargeCenter.quickTitle') }}</p>
          <div class="mt-4 flex flex-wrap gap-2">
            <button
              v-if="paymentEnabled"
              type="button"
              class="recharge-quick-link"
              @click="openOrders"
            >
              {{ t('rechargeCenter.quickOrders') }}
            </button>
            <button type="button" class="recharge-quick-link" @click="openRedeem">
              {{ t('rechargeCenter.quickRedeem') }}
            </button>
            <a
              v-if="docUrl"
              :href="docUrl"
              target="_blank"
              rel="noopener noreferrer"
              class="recharge-quick-link"
            >
              {{ t('rechargeCenter.quickDocs') }}
            </a>
          </div>
        </div>

        <div class="recharge-panel">
          <p class="recharge-panel-title">{{ t('rechargeCenter.tipTitle') }}</p>
          <p class="mt-3 text-sm leading-6 text-slate-600 dark:text-slate-300">
            {{ tipText }}
          </p>
        </div>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { useAppStore } from '@/stores'
import { useAuthStore } from '@/stores/auth'

const DEFAULT_SHOP_URL = 'https://pay.ldxp.cn/shop/HQP8RZ4F'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const appStore = useAppStore()
const authStore = useAuthStore()

const embeddedMode = computed(() => {
  if (route.query.ui_mode === 'embedded') return true
  if (typeof window === 'undefined') return false
  return window.self !== window.top
})

const siteName = computed(() => appStore.siteName || appStore.cachedPublicSettings?.site_name || 'Bridgemind')
const docUrl = computed(() => appStore.docUrl || appStore.cachedPublicSettings?.doc_url || '')
const paymentEnabled = computed(() => appStore.cachedPublicSettings?.payment_enabled ?? false)
const contactLabel = computed(() => {
  const raw = appStore.contactInfo || appStore.cachedPublicSettings?.contact_info || ''
  return raw.trim()
})

const userLabel = computed(() => authStore.user?.email || t('rechargeCenter.unknownUser'))
const balanceLabel = computed(() => `$${(authStore.user?.balance ?? 0).toFixed(2)}`)
const paymentStatusLabel = computed(() =>
  paymentEnabled.value ? t('rechargeCenter.statusEnabled') : t('rechargeCenter.statusFallback')
)

const helperText = computed(() =>
  paymentEnabled.value
    ? t('rechargeCenter.helperEnabled')
    : t('rechargeCenter.helperFallback')
)

const tipText = computed(() =>
  paymentEnabled.value
    ? t('rechargeCenter.tipEnabled')
    : t('rechargeCenter.tipFallback')
)

const secondaryBadge = computed(() =>
  paymentEnabled.value ? t('rechargeCenter.direct.badge') : t('rechargeCenter.redeem.badge')
)
const secondaryTitle = computed(() =>
  paymentEnabled.value ? t('rechargeCenter.direct.title') : t('rechargeCenter.redeem.title')
)
const secondaryDescription = computed(() =>
  paymentEnabled.value
    ? t('rechargeCenter.direct.description')
    : t('rechargeCenter.redeem.description')
)
const secondaryCta = computed(() =>
  paymentEnabled.value ? t('rechargeCenter.direct.cta') : t('rechargeCenter.redeem.cta')
)

const summaryItems = computed(() => [
  {
    index: '01',
    label: t('rechargeCenter.summaryOneTitle'),
    desc: t('rechargeCenter.summaryOneDesc'),
  },
  {
    index: '02',
    label: paymentEnabled.value
      ? t('rechargeCenter.summaryTwoTitleEnabled')
      : t('rechargeCenter.summaryTwoTitleFallback'),
    desc: paymentEnabled.value
      ? t('rechargeCenter.summaryTwoDescEnabled')
      : t('rechargeCenter.summaryTwoDescFallback'),
  },
  {
    index: '03',
    label: t('rechargeCenter.summaryThreeTitle'),
    desc: t('rechargeCenter.summaryThreeDesc'),
  },
])

const sourceUrl = computed(() => {
  const raw = typeof route.query.src_url === 'string' ? route.query.src_url : ''
  return sanitizeURL(raw, '')
})

const shopUrl = computed(() => {
  const raw = typeof route.query.shop_url === 'string' ? route.query.shop_url : DEFAULT_SHOP_URL
  return sanitizeURL(raw, DEFAULT_SHOP_URL)
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

function navigateBack() {
  if (sourceUrl.value) {
    navigateToURL(sourceUrl.value)
    return
  }
  navigateInternal('/dashboard')
}

function navigateInternal(path: string) {
  if (typeof window !== 'undefined' && embeddedMode.value && window.top && window.top !== window) {
    window.top.location.href = new URL(path, window.location.origin).toString()
    return
  }
  router.push(path)
}

function navigateToURL(url: string) {
  if (typeof window !== 'undefined' && embeddedMode.value && window.top && window.top !== window) {
    window.top.location.href = url
    return
  }
  window.location.href = url
}

function openSecondaryChannel() {
  navigateInternal(paymentEnabled.value ? '/purchase' : '/redeem')
}

function openOrders() {
  navigateInternal('/orders')
}

function openRedeem() {
  navigateInternal('/redeem')
}

onMounted(async () => {
  if (!appStore.publicSettingsLoaded) {
    await appStore.fetchPublicSettings()
  }
})
</script>

<style scoped>
.recharge-center-shell {
  position: relative;
  overflow: hidden;
  background:
    radial-gradient(circle at top left, rgba(34, 197, 94, 0.12), transparent 34%),
    radial-gradient(circle at bottom right, rgba(14, 165, 233, 0.14), transparent 32%),
    linear-gradient(145deg, #f7fafc 0%, #edf4ff 48%, #f8fbff 100%);
}

.dark .recharge-center-shell {
  background:
    radial-gradient(circle at top left, rgba(34, 197, 94, 0.18), transparent 32%),
    radial-gradient(circle at bottom right, rgba(14, 165, 233, 0.18), transparent 30%),
    linear-gradient(145deg, #08111f 0%, #0c1729 48%, #111827 100%);
}

.recharge-center-standalone {
  min-height: 100vh;
  padding: 1.5rem;
}

.recharge-center-embedded {
  min-height: 100%;
  padding: 1rem;
}

.recharge-center-orb {
  position: absolute;
  border-radius: 9999px;
  filter: blur(60px);
  opacity: 0.55;
}

.recharge-center-orb-left {
  top: -8rem;
  left: -5rem;
  height: 14rem;
  width: 14rem;
  background: rgba(14, 165, 233, 0.22);
}

.recharge-center-orb-right {
  right: -6rem;
  bottom: -8rem;
  height: 16rem;
  width: 16rem;
  background: rgba(34, 197, 94, 0.18);
}

.recharge-hero,
.recharge-panel,
.recharge-card,
.recharge-summary {
  border: 1px solid rgba(148, 163, 184, 0.22);
  background: rgba(255, 255, 255, 0.78);
  backdrop-filter: blur(18px);
  box-shadow: 0 18px 50px rgba(15, 23, 42, 0.08);
}

.dark .recharge-hero,
.dark .recharge-panel,
.dark .recharge-card,
.dark .recharge-summary {
  border-color: rgba(148, 163, 184, 0.16);
  background: rgba(15, 23, 42, 0.72);
  box-shadow: 0 18px 50px rgba(2, 6, 23, 0.35);
}

.recharge-hero,
.recharge-summary,
.recharge-panel,
.recharge-card {
  border-radius: 1.75rem;
}

.recharge-hero {
  padding: 1.5rem;
}

.recharge-summary {
  padding: 1.25rem;
}

.recharge-summary-kicker,
.recharge-panel-title,
.recharge-kicker {
  font-size: 0.72rem;
  font-weight: 700;
  letter-spacing: 0.16em;
  text-transform: uppercase;
  color: rgb(37 99 235);
}

.recharge-summary-row {
  display: grid;
  grid-template-columns: auto 1fr;
  gap: 0.875rem;
  align-items: start;
  border-radius: 1rem;
  padding: 0.9rem;
  background: rgba(241, 245, 249, 0.78);
}

.dark .recharge-summary-row {
  background: rgba(15, 23, 42, 0.7);
}

.recharge-summary-index {
  display: inline-flex;
  height: 2rem;
  width: 2rem;
  align-items: center;
  justify-content: center;
  border-radius: 9999px;
  background: linear-gradient(135deg, #0ea5e9, #22c55e);
  color: white;
  font-size: 0.75rem;
  font-weight: 800;
}

.recharge-chip,
.recharge-pill,
.recharge-inline-link,
.recharge-quick-link,
.recharge-card-badge {
  display: inline-flex;
  align-items: center;
  gap: 0.45rem;
  border-radius: 9999px;
  font-size: 0.78rem;
  font-weight: 700;
}

.recharge-chip,
.recharge-inline-link,
.recharge-quick-link {
  border: 1px solid rgba(148, 163, 184, 0.25);
  background: rgba(255, 255, 255, 0.72);
  color: rgb(15 23 42);
  padding: 0.7rem 1rem;
}

.dark .recharge-chip,
.dark .recharge-inline-link,
.dark .recharge-quick-link {
  border-color: rgba(148, 163, 184, 0.18);
  background: rgba(15, 23, 42, 0.65);
  color: rgb(226 232 240);
}

.recharge-chip-muted {
  color: rgb(71 85 105);
}

.dark .recharge-chip-muted {
  color: rgb(148 163 184);
}

.recharge-inline-link,
.recharge-quick-link {
  transition: transform 0.2s ease, border-color 0.2s ease, background 0.2s ease;
}

.recharge-inline-link:hover,
.recharge-quick-link:hover {
  transform: translateY(-1px);
  border-color: rgba(37, 99, 235, 0.35);
}

.recharge-pill {
  padding: 0.8rem 1rem;
  border: 1px solid rgba(148, 163, 184, 0.18);
  background: rgba(248, 250, 252, 0.88);
  color: rgb(15 23 42);
}

.dark .recharge-pill {
  border-color: rgba(148, 163, 184, 0.14);
  background: rgba(15, 23, 42, 0.66);
  color: rgb(226 232 240);
}

.recharge-pill-label {
  color: rgb(100 116 139);
}

.dark .recharge-pill-label {
  color: rgb(148 163 184);
}

.recharge-card {
  position: relative;
  display: flex;
  min-height: 18rem;
  flex-direction: column;
  gap: 1rem;
  overflow: hidden;
  padding: 1.35rem;
  transition: transform 0.22s ease, box-shadow 0.22s ease, border-color 0.22s ease;
}

.recharge-card:hover {
  transform: translateY(-3px);
  box-shadow: 0 24px 60px rgba(15, 23, 42, 0.14);
}

.dark .recharge-card:hover {
  box-shadow: 0 24px 60px rgba(2, 6, 23, 0.45);
}

.recharge-card-primary {
  background:
    linear-gradient(160deg, rgba(255, 244, 184, 0.78), rgba(255, 255, 255, 0.78) 56%),
    rgba(255, 255, 255, 0.76);
}

.dark .recharge-card-primary {
  background:
    linear-gradient(160deg, rgba(234, 179, 8, 0.16), rgba(15, 23, 42, 0.74) 58%),
    rgba(15, 23, 42, 0.72);
}

.recharge-card-secondary {
  background:
    linear-gradient(160deg, rgba(191, 219, 254, 0.56), rgba(255, 255, 255, 0.78) 58%),
    rgba(255, 255, 255, 0.76);
}

.dark .recharge-card-secondary {
  background:
    linear-gradient(160deg, rgba(59, 130, 246, 0.16), rgba(15, 23, 42, 0.74) 58%),
    rgba(15, 23, 42, 0.72);
}

.recharge-card-badge {
  align-self: flex-start;
  padding: 0.45rem 0.8rem;
  background: rgba(15, 23, 42, 0.9);
  color: white;
}

.recharge-card-badge-soft {
  background: rgba(37, 99, 235, 0.12);
  color: rgb(37 99 235);
}

.dark .recharge-card-badge-soft {
  background: rgba(96, 165, 250, 0.18);
  color: rgb(147 197 253);
}

.recharge-card-icon {
  display: inline-flex;
  height: 3.4rem;
  width: 3.4rem;
  align-items: center;
  justify-content: center;
  border-radius: 1.1rem;
  background: rgba(15, 23, 42, 0.88);
  color: white;
}

.recharge-card-icon-secondary {
  background: rgba(37, 99, 235, 0.14);
  color: rgb(37 99 235);
}

.dark .recharge-card-icon-secondary {
  background: rgba(59, 130, 246, 0.2);
  color: rgb(147 197 253);
}

.recharge-card-icon svg {
  height: 1.6rem;
  width: 1.6rem;
}

.recharge-card-footer {
  margin-top: auto;
  display: inline-flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  border-top: 1px solid rgba(148, 163, 184, 0.18);
  padding-top: 1rem;
  font-size: 0.92rem;
  font-weight: 700;
  color: rgb(15 23 42);
}

.dark .recharge-card-footer {
  border-top-color: rgba(148, 163, 184, 0.14);
  color: rgb(226 232 240);
}

.recharge-panel {
  padding: 1.25rem;
}

@media (max-width: 640px) {
  .recharge-center-standalone,
  .recharge-center-embedded {
    padding: 0.75rem;
  }

  .recharge-hero,
  .recharge-summary,
  .recharge-panel,
  .recharge-card {
    border-radius: 1.4rem;
  }

  .recharge-card {
    min-height: 15.5rem;
  }
}
</style>
