import { describe, expect, it } from 'vitest'
import {
  resolveDisabledPublicFeatureRedirect,
  routeNeedsPublicSettings
} from '@/router/publicFeatureGuard'

describe('public feature guard', () => {
  it('marks password reset routes as requiring public settings', () => {
    expect(routeNeedsPublicSettings({ requiresPasswordResetEnabled: true })).toBe(true)
  })

  it('marks email verification routes as requiring public settings', () => {
    expect(routeNeedsPublicSettings({ requiresEmailVerifyEnabled: true })).toBe(true)
  })

  it('ignores routes without public feature flags', () => {
    expect(routeNeedsPublicSettings({})).toBe(false)
  })

  it('redirects forgot-password to login when password reset is disabled', () => {
    expect(
      resolveDisabledPublicFeatureRedirect(
        { requiresPasswordResetEnabled: true },
        { password_reset_enabled: false, email_verify_enabled: false }
      )
    ).toBe('/login')
  })

  it('allows forgot-password when password reset is enabled', () => {
    expect(
      resolveDisabledPublicFeatureRedirect(
        { requiresPasswordResetEnabled: true },
        { password_reset_enabled: true, email_verify_enabled: false }
      )
    ).toBeNull()
  })

  it('redirects email-verify to register when email verification is disabled', () => {
    expect(
      resolveDisabledPublicFeatureRedirect(
        { requiresEmailVerifyEnabled: true },
        { password_reset_enabled: true, email_verify_enabled: false }
      )
    ).toBe('/register')
  })

  it('allows email-verify when email verification is enabled', () => {
    expect(
      resolveDisabledPublicFeatureRedirect(
        { requiresEmailVerifyEnabled: true },
        { password_reset_enabled: false, email_verify_enabled: true }
      )
    ).toBeNull()
  })

  it('fails closed when guarded settings are unavailable', () => {
    expect(resolveDisabledPublicFeatureRedirect({ requiresPasswordResetEnabled: true }, null)).toBe(
      '/login'
    )
    expect(resolveDisabledPublicFeatureRedirect({ requiresEmailVerifyEnabled: true }, null)).toBe(
      '/register'
    )
  })
})
