-- 0004_alert_routing_rules.up.sql — per-webhook alert routing rules (ADR-002).
--
-- Today every enabled webhook receives every alert (push_service.go SendAlert
-- filters only by Enabled). This table stores per-webhook routing criteria
-- (metric / severities / node_id) so a webhook only receives matching alerts.
-- Webhooks with NO rules keep the current "receive everything" behavior.
-- node-group based routing is intentionally deferred (ADR-002 Tier-2): there
-- is no node_groups concept yet, so we persist node_id only.
CREATE TABLE public.alert_routing_rules (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    owner_user_id uuid NOT NULL,
    name character varying(255) NOT NULL,
    enabled boolean DEFAULT true NOT NULL,
    metric character varying(50),
    severities jsonb,
    node_id character varying(255),
    webhook_id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);

ALTER TABLE ONLY public.alert_routing_rules
    ADD CONSTRAINT alert_routing_rules_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.alert_routing_rules
    ADD CONSTRAINT alert_routing_rules_owner_user_id_fkey FOREIGN KEY (owner_user_id) REFERENCES public.users(user_id) ON DELETE CASCADE;

ALTER TABLE ONLY public.alert_routing_rules
    ADD CONSTRAINT alert_routing_rules_webhook_id_fkey FOREIGN KEY (webhook_id) REFERENCES public.webhooks(id) ON DELETE CASCADE;

CREATE INDEX idx_alert_routing_rules_owner_user_id ON public.alert_routing_rules USING btree (owner_user_id);
CREATE INDEX idx_alert_routing_rules_webhook_id ON public.alert_routing_rules USING btree (webhook_id);
CREATE INDEX idx_alert_routing_rules_enabled ON public.alert_routing_rules USING btree (enabled);
