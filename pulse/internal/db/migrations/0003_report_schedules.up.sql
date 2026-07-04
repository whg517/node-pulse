-- 0003_report_schedules.up.sql — server-side report schedules (ADR-001).
--
-- Previously report schedules lived only in the browser localStorage
-- (frontend settingsStore) and the server never executed them. This table
-- makes schedules durable, owned, and queryable by the scheduler job that
-- generates + emails reports on a daily/weekly/monthly cadence.
CREATE TABLE public.report_schedules (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    owner_user_id uuid NOT NULL,
    name character varying(255) NOT NULL,
    frequency character varying(20) DEFAULT 'daily' NOT NULL,
    time_of_day character varying(8) DEFAULT '09:00' NOT NULL,
    node_ids jsonb NOT NULL,
    metrics jsonb NOT NULL,
    format character varying(10) DEFAULT 'csv' NOT NULL,
    recipient_email character varying(255),
    enabled boolean DEFAULT true NOT NULL,
    last_run_at timestamp with time zone,
    next_run_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT report_schedules_frequency_check CHECK (frequency = ANY (ARRAY['daily'::character varying, 'weekly'::character varying, 'monthly'::character varying])),
    CONSTRAINT report_schedules_format_check CHECK (format = ANY (ARRAY['csv'::character varying, 'pdf'::character varying]))
);

ALTER TABLE ONLY public.report_schedules
    ADD CONSTRAINT report_schedules_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.report_schedules
    ADD CONSTRAINT report_schedules_owner_user_id_fkey FOREIGN KEY (owner_user_id) REFERENCES public.users(user_id) ON DELETE CASCADE;

CREATE INDEX idx_report_schedules_owner_user_id ON public.report_schedules USING btree (owner_user_id);
CREATE INDEX idx_report_schedules_next_run_at ON public.report_schedules USING btree (next_run_at);
CREATE INDEX idx_report_schedules_enabled ON public.report_schedules USING btree (enabled);
