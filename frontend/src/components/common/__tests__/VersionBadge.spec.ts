import { flushPromises, mount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

const authStoreState = vi.hoisted(() => ({
  isAdmin: true,
}))

const systemApiMocks = vi.hoisted(() => ({
  performUpdate: vi.fn(),
  restartService: vi.fn(),
}))

const pageReloadMocks = vi.hoisted(() => ({
  reloadPage: vi.fn(),
}))

const appStoreState = vi.hoisted(() => ({
  versionLoading: false,
  currentVersion: '1.0.0',
  latestVersion: '1.1.0',
  hasUpdate: true,
  releaseInfo: null,
  buildType: 'release',
  canAutoUpdate: false,
  updateHint: 'Manual update required',
  fetchVersion: vi.fn(),
  clearVersionCache: vi.fn(),
}))

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
  useAuthStore: () => authStoreState,
  useAppStore: () => appStoreState,
}))

vi.mock('@/api/admin/system', () => ({
  performUpdate: systemApiMocks.performUpdate,
  restartService: systemApiMocks.restartService,
}))

vi.mock('@/utils/page-reload', () => ({
  reloadPage: pageReloadMocks.reloadPage,
}))

import VersionBadge from '../VersionBadge.vue'

describe('VersionBadge', () => {
  beforeEach(() => {
    appStoreState.versionLoading = false
    appStoreState.currentVersion = '1.0.0'
    appStoreState.latestVersion = '1.1.0'
    appStoreState.hasUpdate = true
    appStoreState.releaseInfo = null
    appStoreState.buildType = 'release'
    appStoreState.canAutoUpdate = false
    appStoreState.updateHint = 'Manual update required'
    appStoreState.fetchVersion.mockReset()
    appStoreState.clearVersionCache.mockReset()
    systemApiMocks.performUpdate.mockReset()
    systemApiMocks.restartService.mockReset()
    pageReloadMocks.reloadPage.mockReset()
  })

  afterEach(() => {
    vi.useRealTimers()
    vi.unstubAllGlobals()
  })

  it('shows the manual update hint instead of the online update button when auto-update is unsupported', async () => {
    const wrapper = mount(VersionBadge, {
      props: {
        version: '1.0.0',
      },
      global: {
        stubs: {
          Icon: true,
        },
      },
    })

    await wrapper.find('button').trigger('click')
    await flushPromises()

    expect(appStoreState.fetchVersion).toHaveBeenCalledWith(false)
    expect(wrapper.text()).toContain('Manual update required')
    expect(wrapper.text()).not.toContain('version.updateNow')
  })

  it('labels the compact badge as the current running version', () => {
    const wrapper = mount(VersionBadge, {
      props: {
        version: '1.0.0',
      },
      global: {
        stubs: {
          Icon: true,
        },
      },
    })

    const trigger = wrapper.get('button')

    expect(trigger.text()).toContain('version.currentVersionShort')
    expect(trigger.text()).toContain('v1.0.0')
    expect(trigger.attributes('title')).toContain('version.currentVersion: v1.0.0')
    expect(trigger.attributes('title')).toContain('version.latestVersion: v1.1.0')
  })

  it('triggers the online update directly when the update card is clicked in supported environments', async () => {
    appStoreState.canAutoUpdate = true
    systemApiMocks.performUpdate.mockResolvedValue({
      message: 'updated',
      need_restart: true,
    })

    const wrapper = mount(VersionBadge, {
      props: {
        version: '1.0.0',
      },
      global: {
        stubs: {
          Icon: true,
        },
      },
    })

    await wrapper.find('button').trigger('click')
    await flushPromises()

    await wrapper.get('[data-testid="version-update-card"]').trigger('click')
    await flushPromises()

    expect(systemApiMocks.performUpdate).toHaveBeenCalledTimes(1)
    expect(appStoreState.clearVersionCache).toHaveBeenCalledTimes(1)
  })

  it('falls back to the up-to-date state when the backend reports that no update is available anymore', async () => {
    appStoreState.canAutoUpdate = true
    systemApiMocks.performUpdate.mockRejectedValue({
      status: 409,
      reason: 'NO_UPDATE_AVAILABLE',
      message: 'already running the latest version',
    })

    const wrapper = mount(VersionBadge, {
      props: {
        version: '1.0.0',
      },
      global: {
        stubs: {
          Icon: true,
        },
      },
    })

    await wrapper.find('button').trigger('click')
    await flushPromises()

    await wrapper.get('[data-testid="version-update-card"]').trigger('click')
    await flushPromises()

    expect(appStoreState.clearVersionCache).toHaveBeenCalledTimes(1)
    expect(appStoreState.fetchVersion).toHaveBeenCalledWith(true)
    expect(wrapper.text()).toContain('version.currentRunningVersion')
    expect(wrapper.text()).not.toContain('version.updateFailed')
  })

  it('describes the no-update dropdown state as the current running version', async () => {
    appStoreState.latestVersion = '1.0.0'
    appStoreState.hasUpdate = false

    const wrapper = mount(VersionBadge, {
      props: {
        version: '1.0.0',
      },
      global: {
        stubs: {
          Icon: true,
        },
      },
    })

    await wrapper.find('button').trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('version.currentRunningVersion')
    expect(wrapper.text()).not.toContain('version.upToDate')
  })

  it('waits for service replacement and reloads automatically when the backend asks to poll for restart', async () => {
    vi.useFakeTimers()
    appStoreState.canAutoUpdate = true
    systemApiMocks.performUpdate.mockResolvedValue({
      message: 'update started',
      need_restart: false,
      poll_for_restart: true,
    })

    const fetchMock = vi
      .fn()
      .mockRejectedValueOnce(new Error('service restarting'))
      .mockResolvedValueOnce({ ok: true })
    vi.stubGlobal('fetch', fetchMock)

    const wrapper = mount(VersionBadge, {
      props: {
        version: '1.0.0',
      },
      global: {
        stubs: {
          Icon: true,
        },
      },
    })

    await wrapper.find('button').trigger('click')
    await flushPromises()

    await wrapper.get('[data-testid="version-update-card"]').trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('version.updateStarted')
    expect(fetchMock).toHaveBeenCalledTimes(1)

    await vi.advanceTimersByTimeAsync(5000)
    await flushPromises()

    expect(fetchMock).toHaveBeenCalledTimes(2)
    expect(pageReloadMocks.reloadPage).toHaveBeenCalledTimes(1)
  })

  it('keeps the unsupported update card static instead of turning it into a redirecting link', async () => {
    appStoreState.releaseInfo = {
      html_url: 'https://example.com/releases/v1.1.0',
      name: 'v1.1.0',
      body: 'notes',
      published_at: '2026-04-25T00:00:00Z',
    }

    const wrapper = mount(VersionBadge, {
      props: {
        version: '1.0.0',
      },
      global: {
        stubs: {
          Icon: true,
        },
      },
    })

    await wrapper.find('button').trigger('click')
    await flushPromises()

    expect(wrapper.get('[data-testid="version-update-card"]').element.tagName).toBe('DIV')
    expect(wrapper.get('[data-testid="version-changelog-link"]').attributes('href')).toBe(
      'https://example.com/releases/v1.1.0',
    )
  })

  it('renders git-style update targets without forcing a semantic-version prefix', async () => {
    appStoreState.latestVersion = 'origin/main@95ecf821'

    const wrapper = mount(VersionBadge, {
      props: {
        version: '1.0.0',
      },
      global: {
        stubs: {
          Icon: true,
        },
      },
    })

    await wrapper.find('button').trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('origin/main@95ecf821')
    expect(wrapper.text()).not.toContain('vorigin/main@95ecf821')
  })
})
