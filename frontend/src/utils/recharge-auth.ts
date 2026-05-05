export function resolveRechargeAuthToken(
  embeddedToken?: string | null,
  storedToken?: string | null,
): string {
  const preferred = String(embeddedToken ?? '').trim()
  if (preferred) return preferred
  return String(storedToken ?? '').trim()
}

export function buildRechargeAuthHeaders(
  embeddedToken?: string | null,
  storedToken?: string | null,
): Record<string, string> {
  const token = resolveRechargeAuthToken(embeddedToken, storedToken)
  if (!token) return {}
  return {
    Authorization: `Bearer ${token}`,
  }
}

export function canUseEmbeddedRechargeAccess(
  path?: string | null,
  embeddedToken?: string | null,
): boolean {
  const normalizedPath = String(path ?? '').trim()
  const token = String(embeddedToken ?? '').trim()
  if (!normalizedPath || !token) return false
  return normalizedPath === '/recharge-center' || normalizedPath.startsWith('/custom/')
}

export function appendRechargeAccessToken(
  path?: string | null,
  embeddedToken?: string | null,
): string {
  const normalizedPath = String(path ?? '').trim()
  const token = String(embeddedToken ?? '').trim()
  if (!normalizedPath || !token) return normalizedPath

  let parsed: URL
  try {
    parsed = new URL(normalizedPath, 'https://embedded.local')
  } catch {
    return normalizedPath
  }

  if (!canUseEmbeddedRechargeAccess(parsed.pathname, token)) {
    return normalizedPath
  }

  if (!parsed.searchParams.get('token')) {
    parsed.searchParams.set('token', token)
  }

  return `${parsed.pathname}${parsed.search}${parsed.hash}`
}
