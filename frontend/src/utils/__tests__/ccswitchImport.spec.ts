import { describe, expect, it } from 'vitest'
import { buildCcsImportDeeplink, decodeCcsBase64JsonParam } from '../ccswitchImport'

const sanitizeLikeCurrentCcs = (name: string) => {
  const key = name
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9_]/g, '_')
    .replace(/^_+|_+$/g, '')
  return key || 'custom'
}

const simulateCurrentCcsCodexImportConfig = (url: URL) => {
  const providerName = url.searchParams.get('name') || 'custom'
  const providerKey = sanitizeLikeCurrentCcs(providerName)
  const model = url.searchParams.get('model') || 'gpt-5-codex'
  const endpoint = (url.searchParams.get('endpoint') || '').replace(/\/+$/, '')

  return `model_provider = "${providerKey}"
model = "${model}"
model_reasoning_effort = "high"
disable_response_storage = true

[model_providers.${providerKey}]
name = "${providerKey}"
base_url = "${endpoint}"
wire_api = "responses"
requires_openai_auth = true
`
}

describe('ccswitchImport', () => {
  it('builds a mixed Codex import with visible multi-model profiles', () => {
    const deeplink = buildCcsImportDeeplink({
      apiKey: 'sk-test',
      baseUrl: 'https://bridgemind.pro/v1',
      providerName: 'BridgeMind',
      platform: 'mixed',
      clientType: 'codex',
    })

    const url = new URL(deeplink)
    expect(url.protocol).toBe('ccswitch:')
    expect(url.searchParams.get('app')).toBe('codex')
    expect(url.searchParams.get('name')).toBe('BridgeMind')
    expect(url.searchParams.get('homepage')).toBe('https://bridgemind.pro/v1')
    expect(url.searchParams.get('endpoint')).toBe('https://bridgemind.pro/v1')
    expect(url.searchParams.get('model')).toContain('gpt-5.5')
    expect(url.searchParams.get('model')).toContain('[profiles.claude]')
    expect(url.searchParams.get('model')).toContain('[profiles.gemini]')
    expect(url.searchParams.get('model')).toContain('[profiles.deepseek]')
    expect(url.searchParams.get('notes')).toContain('claude=claude-sonnet-4-6')
    expect(url.searchParams.get('notes')).toContain('gemini=gemini-2.5-pro')
    expect(url.searchParams.get('notes')).toContain('deepseek=deepseek-v4-pro')

    const config = decodeCcsBase64JsonParam<{ auth: Record<string, string>; config: string }>(
      url.searchParams.get('config') || '',
    )
    expect(config.auth.OPENAI_API_KEY).toBe('sk-test')
    expect(config.config).toContain('[profiles.gpt]')
    expect(config.config).toContain('model = "gpt-5.5"')
    expect(config.config).toContain('[profiles.codex]')
    expect(config.config).toContain('model = "gpt-5.3-codex"')
    expect(config.config).toContain('[profiles.claude]')
    expect(config.config).toContain('model = "claude-sonnet-4-6"')
    expect(config.config).toContain('[profiles.gemini]')
    expect(config.config).toContain('model = "gemini-2.5-pro"')
    expect(config.config).toContain('[profiles.deepseek]')
    expect(config.config).toContain('model = "deepseek-v4-pro"')
    expect(config.config).toContain('[model_providers.bridgemind]')
    expect(config.config).toContain('base_url = "https://bridgemind.pro/v1"')
  })

  it('keeps mixed Codex profiles after current CC-Switch rebuilds config.toml', () => {
    const deeplink = buildCcsImportDeeplink({
      apiKey: 'sk-test',
      baseUrl: 'https://bridgemind.pro/v1',
      providerName: 'BridgeMind',
      platform: 'mixed',
      clientType: 'codex',
    })

    const rebuiltConfig = simulateCurrentCcsCodexImportConfig(new URL(deeplink))
    expect(rebuiltConfig).toContain('model_provider = "bridgemind"')
    expect(rebuiltConfig).toContain('base_url = "https://bridgemind.pro/v1"')
    expect(rebuiltConfig).toContain('[profiles.gpt]')
    expect(rebuiltConfig).toContain('model = "gpt-5.5"')
    expect(rebuiltConfig).toContain('[profiles.codex]')
    expect(rebuiltConfig).toContain('model = "gpt-5.3-codex"')
    expect(rebuiltConfig).toContain('[profiles.claude]')
    expect(rebuiltConfig).toContain('model = "claude-sonnet-4-6"')
    expect(rebuiltConfig).toContain('[profiles.gemini]')
    expect(rebuiltConfig).toContain('model = "gemini-2.5-pro"')
    expect(rebuiltConfig).toContain('[profiles.deepseek]')
    expect(rebuiltConfig).toContain('model = "deepseek-v4-pro"')
  })

  it('routes mixed DeepSeek imports to OpenCode instead of Codex responses', () => {
    const deeplink = buildCcsImportDeeplink({
      apiKey: 'sk-test',
      baseUrl: 'https://bridgemind.pro/v1',
      providerName: 'BridgeMind',
      platform: 'mixed',
      clientType: 'deepseek',
    })

    const url = new URL(deeplink)
    expect(url.searchParams.get('app')).toBe('opencode')
    expect(url.searchParams.get('model')).toBe('deepseek-v4-pro')
    expect(url.searchParams.get('config')).toBeNull()
    expect(url.searchParams.get('notes')).toContain('OpenAI-compatible chat')
  })

  it('keeps Antigravity client imports on the antigravity endpoint', () => {
    const deeplink = buildCcsImportDeeplink({
      apiKey: 'sk-test',
      baseUrl: 'https://bridgemind.pro/v1',
      providerName: 'BridgeMind',
      platform: 'antigravity',
      clientType: 'gemini',
    })

    const url = new URL(deeplink)
    expect(url.searchParams.get('app')).toBe('gemini')
    expect(url.searchParams.get('endpoint')).toBe('https://bridgemind.pro/v1/antigravity')
    expect(url.searchParams.get('model')).toBe('gemini-2.5-pro')
  })
})
