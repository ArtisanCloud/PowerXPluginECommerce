package subscription_reconciliation

import (
	"context"
	"encoding/json"
	"strings"
	"time"
)

var defaultRetryWindows = []string{"1h", "24h", "72h", "7d"}

type governancePolicy struct {
	retryWindows        []time.Duration
	retryWindowLabels   []string
	escalationThreshold int
}

func (s *Service) loadGovernancePolicy(ctx context.Context) (governancePolicy, error) {
	policy := governancePolicy{
		retryWindows:        parseRetryWindows(defaultRetryWindows),
		retryWindowLabels:   append([]string(nil), defaultRetryWindows...),
		escalationThreshold: 3,
	}
	row, err := s.policyRepo.GetLatestEnabled(ctx)
	if err != nil {
		return governancePolicy{}, err
	}
	if row == nil {
		return policy, nil
	}
	if row.EscalationThreshold > 0 {
		policy.escalationThreshold = row.EscalationThreshold
	}
	if len(row.RetryWindows) > 0 {
		labels := make([]string, 0, len(defaultRetryWindows))
		if err := json.Unmarshal(row.RetryWindows, &labels); err == nil {
			labels = compactStrings(labels)
			if len(labels) > 0 {
				parsed := parseRetryWindows(labels)
				if len(parsed) > 0 {
					policy.retryWindowLabels = labels
					policy.retryWindows = parsed
				}
			}
		}
	}
	return policy, nil
}

func compactStrings(values []string) []string {
	out := make([]string, 0, len(values))
	for _, v := range values {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		out = append(out, v)
	}
	return out
}

func parseRetryWindows(values []string) []time.Duration {
	out := make([]time.Duration, 0, len(values))
	for _, v := range values {
		d, ok := parseRetryWindow(v)
		if !ok {
			continue
		}
		out = append(out, d)
	}
	return out
}

func parseRetryWindow(raw string) (time.Duration, bool) {
	v := strings.TrimSpace(strings.ToLower(raw))
	switch v {
	case "1h":
		return time.Hour, true
	case "24h":
		return 24 * time.Hour, true
	case "72h":
		return 72 * time.Hour, true
	case "7d":
		return 7 * 24 * time.Hour, true
	default:
		return 0, false
	}
}
