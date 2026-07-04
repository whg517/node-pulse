package webhook

import (
	"context"

	"github.com/whg517/node-pulse/pulse/internal/models"
)

// RuleStore is the minimal subset of AlertRoutingRulesRepository the matcher
// needs. Defined locally to keep webhook free of a db import cycle coupling.
type RuleStore interface {
	ListEnabled(ctx context.Context) ([]*models.AlertRoutingRule, error)
}

// ruleRouter implements RouteMatcher using persisted AlertRoutingRules.
//
// Semantics (ADR-002 Tier-1):
//   - A webhook with NO enabled rules is always kept (receive everything).
//   - A webhook WITH rules is kept only if at least one rule matches the event.
//   - A rule matches when all of its non-empty criteria match (AND across
//     criteria; null criterion = wildcard). Severity matches if the event's
//     Level is in the rule's Severities list (when that list is non-empty).
type ruleRouter struct {
	store RuleStore
}

// NewRuleRouter builds a RouteMatcher backed by the given rule store.
func NewRuleRouter(store RuleStore) RouteMatcher {
	if store == nil {
		return nil
	}
	return &ruleRouter{store: store}
}

func (r *ruleRouter) Filter(ctx context.Context, event *models.AlertEvent, candidates []*models.Webhook) []*models.Webhook {
	rules, err := r.store.ListEnabled(ctx)
	if err != nil || len(rules) == 0 {
		// No rules at all (or query failed) => legacy behavior: keep all.
		return candidates
	}

	// Group enabled rules by webhook ID.
	rulesByWebhook := make(map[string][]*models.AlertRoutingRule, len(rules))
	for _, rule := range rules {
		rulesByWebhook[rule.WebhookID] = append(rulesByWebhook[rule.WebhookID], rule)
	}

	out := make([]*models.Webhook, 0, len(candidates))
	for _, wh := range candidates {
		whRules := rulesByWebhook[wh.ID]
		if len(whRules) == 0 {
			// No rules for this webhook => keep (receive everything).
			out = append(out, wh)
			continue
		}
		for _, rule := range whRules {
			if ruleMatches(rule, event) {
				out = append(out, wh)
				break // one matching rule is enough; don't double-send
			}
		}
	}
	return out
}

func ruleMatches(rule *models.AlertRoutingRule, event *models.AlertEvent) bool {
	if rule.Metric != "" && rule.Metric != event.Metric {
		return false
	}
	if len(rule.Severities) > 0 {
		matched := false
		for _, sev := range rule.Severities {
			if sev == event.Level {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	if rule.NodeID != "" && rule.NodeID != event.NodeID {
		return false
	}
	return true
}
