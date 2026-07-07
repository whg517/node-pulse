-- Reverse 0006: drop custom_headers from webhooks.
ALTER TABLE public.webhooks
    DROP COLUMN IF EXISTS custom_headers;
