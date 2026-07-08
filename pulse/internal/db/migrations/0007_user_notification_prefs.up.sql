-- 0007: Per-user notification preferences (F4 Phase 2).
-- Stores each user's choice of which alert severities trigger an email
-- notification. NULL row = defaults (email_enabled true, min_alert_level 'P1').
-- One row per user (user_id is PK).
CREATE TABLE IF NOT EXISTS public.user_notification_prefs (
    user_id         uuid          PRIMARY KEY REFERENCES public.users(user_id) ON DELETE CASCADE,
    email_enabled   boolean       NOT NULL DEFAULT true,
    -- Minimum alert severity that triggers an email. P0 = critical only,
    -- P1 = warnings + critical, P2 = all. Mirrors the client-side F4 Phase 1
    -- severity floor so the two compose (a user can set a loose server-side
    -- floor and a tighter client-side browser filter, or vice versa).
    min_alert_level varchar(8)    NOT NULL DEFAULT 'P1',
    -- Optional override destination; NULL = use the user's profile email.
    notify_email    varchar(255),
    updated_at      timestamp with time zone DEFAULT now()
);
