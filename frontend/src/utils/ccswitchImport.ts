import type { GroupPlatform } from '@/types'

export type CcsClientType = 'claude' | 'gemini' | 'codex' | 'deepseek'

interface BuildCcsImportOptions {
  apiKey: string
  baseUrl: string
  providerName: string
  platform: GroupPlatform | null | undefined
  clientType: CcsClientType
}

interface CcsTargetConfig {
  app: 'claude' | 'gemini' | 'codex' | 'opencode'
  name?: string
  homepage?: string
  endpoint: string
  model?: string
  icon?: string
  notes?: string
  extraParams?: Record<string, string>
  config?: string
}

const mixedCodexProfiles = {
  gpt: 'gpt-5.5',
  codex: 'gpt-5.3-codex',
  claude: 'claude-sonnet-4-6',
  gemini: 'gemini-2.5-pro',
  deepseek: 'deepseek-v4-pro',
} as const

const defaultModels = {
  codex: mixedCodexProfiles.gpt,
  claude: mixedCodexProfiles.claude,
  gemini: mixedCodexProfiles.gemini,
  deepseek: mixedCodexProfiles.deepseek,
} as const

const trimTrailingSlash = (value: string) => value.replace(/\/+$/, '')

const encodeBase64Utf8 = (value: string) => {
  const encoder = typeof TextEncoder !== 'undefined' ? new TextEncoder() : null
  if (!encoder) {
    return btoa(unescape(encodeURIComponent(value)))
  }

  const bytes = encoder.encode(value)
  let binary = ''
  bytes.forEach((byte) => {
    binary += String.fromCharCode(byte)
  })
  return btoa(binary)
}

const sanitizeCodexProviderName = (name: string) => {
  const key = name
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9_]/g, '_')
    .replace(/^_+|_+$/g, '')
  return key || 'bridgemind'
}

const escapeTomlString = (value: string) => value.replace(/\\/g, '\\\\').replace(/"/g, '\\"')

const buildUsageScript = () => `({
  request: {
    url: "{{baseUrl}}/v1/usage",
    method: "GET",
    headers: { "Authorization": "Bearer {{apiKey}}" }
  },
  extractor: function(response) {
    const remaining = response?.remaining ?? response?.quota?.remaining ?? response?.balance;
    const unit = response?.unit ?? response?.quota?.unit ?? "USD";
    return {
      isValid: response?.is_active ?? response?.isValid ?? true,
      remaining,
      unit
    };
  }
})`

const buildMixedCodexConfig = (providerName: string, endpoint: string, apiKey: string) => {
  const providerKey = sanitizeCodexProviderName(providerName)
  const safeProviderKey = escapeTomlString(providerKey)
  const safeProviderName = escapeTomlString(providerName)
  const safeEndpoint = escapeTomlString(trimTrailingSlash(endpoint))

  const config = `model_provider = "${safeProviderKey}"
model = "${mixedCodexProfiles.gpt}"
review_model = "${mixedCodexProfiles.gpt}"
model_reasoning_effort = "high"
disable_response_storage = true
network_access = "enabled"
model_context_window = 1000000
model_auto_compact_token_limit = 900000

[profiles.gpt]
model_provider = "${safeProviderKey}"
model = "${mixedCodexProfiles.gpt}"
model_reasoning_effort = "high"

[profiles.codex]
model_provider = "${safeProviderKey}"
model = "${mixedCodexProfiles.codex}"
model_reasoning_effort = "high"

[profiles.claude]
model_provider = "${safeProviderKey}"
model = "${mixedCodexProfiles.claude}"
model_reasoning_effort = "high"

[profiles.gemini]
model_provider = "${safeProviderKey}"
model = "${mixedCodexProfiles.gemini}"
model_reasoning_effort = "high"

[profiles.deepseek]
model_provider = "${safeProviderKey}"
model = "${mixedCodexProfiles.deepseek}"
model_reasoning_effort = "high"

[model_providers.${providerKey}]
name = "${safeProviderName}"
base_url = "${safeEndpoint}"
wire_api = "responses"
requires_openai_auth = true`

  return encodeBase64Utf8(JSON.stringify({
    auth: {
      OPENAI_API_KEY: apiKey,
    },
    config,
  }))
}

const buildMixedCodexLegacyCcsModel = (providerName: string) => {
  const providerKey = sanitizeCodexProviderName(providerName)
  const safeProviderKey = escapeTomlString(providerKey)

  // Current CC-Switch rebuilds Codex config.toml from URL params and ignores the
  // embedded config body. Keep profiles visible after that rebuild by embedding
  // the extra TOML in the model field it preserves.
  return `${mixedCodexProfiles.gpt}"
review_model = "${mixedCodexProfiles.gpt}"
model_reasoning_effort = "high"
disable_response_storage = true
network_access = "enabled"
model_context_window = 1000000
model_auto_compact_token_limit = 900000

[profiles.gpt]
model_provider = "${safeProviderKey}"
model = "${mixedCodexProfiles.gpt}"
model_reasoning_effort = "high"

[profiles.codex]
model_provider = "${safeProviderKey}"
model = "${mixedCodexProfiles.codex}"
model_reasoning_effort = "high"

[profiles.claude]
model_provider = "${safeProviderKey}"
model = "${mixedCodexProfiles.claude}"
model_reasoning_effort = "high"

[profiles.gemini]
model_provider = "${safeProviderKey}"
model = "${mixedCodexProfiles.gemini}"
model_reasoning_effort = "high"

[profiles.deepseek]
model_provider = "${safeProviderKey}"
model = "${mixedCodexProfiles.deepseek}"
#`
}

const buildPlainCodexConfig = (
  providerName: string,
  endpoint: string,
  apiKey: string,
  model: string,
) => {
  const providerKey = sanitizeCodexProviderName(providerName)
  const safeProviderKey = escapeTomlString(providerKey)
  const safeProviderName = escapeTomlString(providerName)
  const safeEndpoint = escapeTomlString(trimTrailingSlash(endpoint))

  const config = `model_provider = "${safeProviderKey}"
model = "${escapeTomlString(model)}"
model_reasoning_effort = "high"
disable_response_storage = true
network_access = "enabled"

[model_providers.${providerKey}]
name = "${safeProviderName}"
base_url = "${safeEndpoint}"
wire_api = "responses"
requires_openai_auth = true`

  return encodeBase64Utf8(JSON.stringify({
    auth: {
      OPENAI_API_KEY: apiKey,
    },
    config,
  }))
}

const resolveCcsTarget = ({
  apiKey,
  baseUrl,
  providerName,
  platform,
  clientType,
}: BuildCcsImportOptions): CcsTargetConfig => {
  const normalizedBaseUrl = trimTrailingSlash(baseUrl)

  if (platform === 'antigravity') {
    const endpoint = `${normalizedBaseUrl}/antigravity`
    if (clientType === 'gemini') {
      return {
        app: 'gemini',
        endpoint,
        model: defaultModels.gemini,
        icon: 'gemini',
      }
    }

    return {
      app: 'claude',
      endpoint,
      model: defaultModels.claude,
      icon: 'claude',
      extraParams: {
        sonnetModel: defaultModels.claude,
        opusModel: 'claude-opus-4-6-thinking',
      },
    }
  }

  if (clientType === 'gemini' || platform === 'gemini') {
    return {
      app: 'gemini',
      endpoint: normalizedBaseUrl,
      model: defaultModels.gemini,
      icon: 'gemini',
    }
  }

  if (clientType === 'claude' || platform === 'anthropic') {
    return {
      app: 'claude',
      endpoint: normalizedBaseUrl,
      model: defaultModels.claude,
      icon: 'claude',
      extraParams: {
        sonnetModel: defaultModels.claude,
        opusModel: 'claude-opus-4-6-thinking',
      },
    }
  }

  if (clientType === 'deepseek' || platform === 'deepseek') {
    return {
      app: 'opencode',
      endpoint: normalizedBaseUrl,
      model: defaultModels.deepseek,
      icon: 'deepseek',
      notes: 'DeepSeek uses the OpenAI-compatible chat interface. Codex currently requires Responses API providers.',
    }
  }

  const mixedCodexProviderName = providerName
  const mixedCodexEndpoint = normalizedBaseUrl
  const model = platform === 'mixed'
    ? buildMixedCodexLegacyCcsModel(mixedCodexProviderName)
    : defaultModels.codex
  return {
    app: 'codex',
    name: platform === 'mixed' ? mixedCodexProviderName : undefined,
    homepage: platform === 'mixed' ? mixedCodexEndpoint : undefined,
    endpoint: mixedCodexEndpoint,
    model,
    icon: 'openai',
    notes: platform === 'mixed'
      ? `Codex profiles: gpt=${mixedCodexProfiles.gpt}, codex=${mixedCodexProfiles.codex}, claude=${mixedCodexProfiles.claude}, gemini=${mixedCodexProfiles.gemini}, deepseek=${mixedCodexProfiles.deepseek}.`
      : undefined,
    config: platform === 'mixed'
      ? buildMixedCodexConfig(mixedCodexProviderName, mixedCodexEndpoint, apiKey)
      : buildPlainCodexConfig(providerName, mixedCodexEndpoint, apiKey, model),
  }
}

export const buildCcsImportDeeplink = (options: BuildCcsImportOptions) => {
  const target = resolveCcsTarget(options)
  const usageScript = buildUsageScript()

  const params = new URLSearchParams({
    resource: 'provider',
    app: target.app,
    name: target.name || options.providerName,
    homepage: target.homepage || trimTrailingSlash(options.baseUrl),
    endpoint: target.endpoint,
    apiKey: options.apiKey,
    configFormat: 'json',
    usageEnabled: 'true',
    usageScript: encodeBase64Utf8(usageScript),
    usageAutoInterval: '30',
  })

  if (target.model) params.set('model', target.model)
  if (target.icon) params.set('icon', target.icon)
  if (target.notes) params.set('notes', target.notes)
  if (target.config) params.set('config', target.config)
  Object.entries(target.extraParams || {}).forEach(([key, value]) => {
    params.set(key, value)
  })

  return `ccswitch://v1/import?${params.toString()}`
}

export const decodeCcsBase64JsonParam = <T = unknown>(value: string): T => {
  const binary = atob(value)
  const bytes = Uint8Array.from(binary, (char) => char.charCodeAt(0))
  const decoded = new TextDecoder().decode(bytes)
  return JSON.parse(decoded) as T
}
