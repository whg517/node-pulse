-- 0002_export_tasks.up.sql — persist data export tasks across server restarts.
--
-- Background: ExportTask previously lived only in process memory
-- (internal/export/service.go `tasks map`). A server restart lost all running
-- and historical export tasks. This table durably stores export tasks so the
-- service can recover pending work and serve history queries after restart.
--
-- Convention follows 0001_init.up.sql (uuid PK, CHECK constraint for enum-like
-- status, jsonb for arrays, FK with ON DELETE CASCADE, idx_<table>_<col>).
CREATE TABLE public.export_tasks (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    node_ids jsonb NOT NULL,
    start_time timestamp with time zone NOT NULL,
    end_time timestamp with time zone NOT NULL,
    metrics jsonb NOT NULL,
    format character varying(10) DEFAULT 'csv' NOT NULL,
    status character varying(20) DEFAULT 'pending' NOT NULL,
    file_path text,
    file_size bigint DEFAULT 0 NOT NULL,
    record_count integer DEFAULT 0 NOT NULL,
    error_message text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    completed_at timestamp with time zone,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT export_tasks_status_check CHECK (status = ANY (ARRAY['pending'::character varying, 'processing'::character varying, 'completed'::character varying, 'failed'::character varying])),
    CONSTRAINT export_tasks_format_check CHECK (format = ANY (ARRAY['csv'::character varying, 'xlsx'::character varying]))
);

ALTER TABLE ONLY public.export_tasks
    ADD CONSTRAINT export_tasks_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.export_tasks
    ADD CONSTRAINT export_tasks_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON DELETE CASCADE;

CREATE INDEX idx_export_tasks_user_id ON public.export_tasks USING btree (user_id);
CREATE INDEX idx_export_tasks_status ON public.export_tasks USING btree (status);
CREATE INDEX idx_export_tasks_created_at ON public.export_tasks USING btree (created_at DESC);
