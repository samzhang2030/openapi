-- mixed-default is the automatic target for OpenAI API key accounts.
-- It must be an OpenAI group so /v1/responses routes to OpenAI scheduling.
UPDATE groups
SET platform = 'openai',
    updated_at = NOW()
WHERE name = 'mixed-default'
  AND platform = 'mixed'
  AND deleted_at IS NULL;
