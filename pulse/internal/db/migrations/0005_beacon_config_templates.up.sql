-- 0005_beacon_config_templates.up.sql — shared Beacon config templates (ADR-003).
--
-- Previously config templates lived only in browser localStorage. This table
-- makes templates server-owned, shareable across the team, and reusable via
-- the existing "apply template" flow (which calls POST /beacons/:id/config).
CREATE TABLE public.beacon_config_templates (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    owner_user_id uuid NOT NULL,
    name character varying(255) NOT NULL,
    description text,
    probes jsonb NOT NULL,
    interval_seconds integer NOT NULL,
    timeout_seconds integer NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);

ALTER TABLE ONLY public.beacon_config_templates
    ADD CONSTRAINT beacon_config_templates_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.beacon_config_templates
    ADD CONSTRAINT beacon_config_templates_owner_user_id_fkey FOREIGN KEY (owner_user_id) REFERENCES public.users(user_id) ON DELETE CASCADE;

CREATE INDEX idx_beacon_config_templates_owner_user_id ON public.beacon_config_templates USING btree (owner_user_id);
