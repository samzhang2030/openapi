import type { RouteMeta } from 'vue-router'
import type { PublicSettings } from '@/types'

type PublicFeatureRouteMeta = Pick<
  RouteMeta,
  'requiresPasswordResetEnabled' | 'requiresEmailVerifyEnabled'
>

type PublicFeatureSettings = Pick<
  PublicSettings,
  'password_reset_enabled' | 'email_verify_enabled'
>

export function routeNeedsPublicSettings(meta: PublicFeatureRouteMeta): boolean {
  return meta.requiresPasswordResetEnabled === true || meta.requiresEmailVerifyEnabled === true
}

export function resolveDisabledPublicFeatureRedirect(
  meta: PublicFeatureRouteMeta,
  settings: PublicFeatureSettings | null | undefined
): string | null {
  if (meta.requiresPasswordResetEnabled && settings?.password_reset_enabled !== true) {
    return '/login'
  }

  if (meta.requiresEmailVerifyEnabled && settings?.email_verify_enabled !== true) {
    return '/register'
  }

  return null
}
