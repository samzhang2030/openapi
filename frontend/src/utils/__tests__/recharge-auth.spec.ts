import { describe, expect, it } from 'vitest'
import {
  appendRechargeAccessToken,
  buildRechargeAuthHeaders,
  canUseEmbeddedRechargeAccess,
  resolveRechargeAuthToken,
} from '../recharge-auth'

describe('recharge-auth', () => {
  it('prefers embedded token over local storage token', () => {
    expect(resolveRechargeAuthToken('embedded-token', 'local-token')).toBe('embedded-token')
  })

  it('falls back to local storage token when embedded token is empty', () => {
    expect(resolveRechargeAuthToken('  ', 'local-token')).toBe('local-token')
  })

  it('builds bearer authorization headers when a token is available', () => {
    expect(buildRechargeAuthHeaders('embedded-token', null)).toEqual({
      Authorization: 'Bearer embedded-token',
    })
  })

  it('returns an empty header object when no token is available', () => {
    expect(buildRechargeAuthHeaders('', '')).toEqual({})
  })

  it('allows tokenized recharge center routes to bypass login gating', () => {
    expect(canUseEmbeddedRechargeAccess('/recharge-center', 'embedded-token')).toBe(true)
    expect(canUseEmbeddedRechargeAccess('/custom/da51f3671c989d4b', 'embedded-token')).toBe(true)
  })

  it('does not allow bypass when token or path is missing', () => {
    expect(canUseEmbeddedRechargeAccess('/recharge-center', '')).toBe(false)
    expect(canUseEmbeddedRechargeAccess('/dashboard', 'embedded-token')).toBe(false)
  })

  it('appends the embedded token to internal recharge paths that would otherwise lose auth', () => {
    expect(
      appendRechargeAccessToken('/recharge-center?shop_url=https%3A%2F%2Fpay.ldxp.cn%2Fshop%2FHQP8RZ4F', 'embedded-token'),
    ).toBe('/recharge-center?shop_url=https%3A%2F%2Fpay.ldxp.cn%2Fshop%2FHQP8RZ4F&token=embedded-token')

    expect(
      appendRechargeAccessToken('/custom/da51f3671c989d4b', 'embedded-token'),
    ).toBe('/custom/da51f3671c989d4b?token=embedded-token')
  })

  it('does not overwrite existing tokens or touch unrelated paths', () => {
    expect(
      appendRechargeAccessToken('/recharge-center?token=kept-token', 'embedded-token'),
    ).toBe('/recharge-center?token=kept-token')
    expect(
      appendRechargeAccessToken('/dashboard', 'embedded-token'),
    ).toBe('/dashboard')
  })
})
