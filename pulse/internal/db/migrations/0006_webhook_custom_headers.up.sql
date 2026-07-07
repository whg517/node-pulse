-- 0006: Add custom_headers to webhooks (J9).
-- Stores a JSON object of extra HTTP headers to send with every delivery,
-- e.g. {"Authorization": "Bearer xxx", "X-Tenant": "ops"}. NULL = none.
-- Kept separate from event_format (which is the payload template) so the
-- two concerns stay independent.
ALTER TABLE public.webhooks
    ADD COLUMN IF NOT EXISTS custom_headers jsonb;
