package promotion

import (
	"sort"
	"strings"
)

type quoteCandidate struct {
	Applied  AppliedPromotion
	Stacking StackingRule
}

func selectApplicablePromotions(candidates []quoteCandidate) ([]AppliedPromotion, []RejectedPromotion) {
	if len(candidates) == 0 {
		return []AppliedPromotion{}, []RejectedPromotion{}
	}
	byGroup := make(map[string]quoteCandidate)
	var ungrouped []quoteCandidate
	var rejected []RejectedPromotion
	for _, candidate := range candidates {
		groups := exclusionGroups(candidate.Stacking)
		if len(groups) == 0 {
			ungrouped = append(ungrouped, candidate)
			continue
		}
		current, conflict := bestGroupConflict(byGroup, groups)
		if conflict && current.Applied.DiscountMinor >= candidate.Applied.DiscountMinor {
			rejected = append(rejected, RejectedPromotion{
				PromotionID: candidate.Applied.PromotionID,
				Code:        candidate.Applied.Code,
				Reason:      ReasonExclusiveConflict,
			})
			continue
		}
		if conflict {
			rejected = append(rejected, RejectedPromotion{
				PromotionID: current.Applied.PromotionID,
				Code:        current.Applied.Code,
				Reason:      ReasonExclusiveConflict,
			})
			removeCandidateFromGroups(byGroup, current)
		}
		for _, group := range groups {
			byGroup[group] = candidate
		}
	}
	selectedMap := make(map[string]quoteCandidate)
	for _, candidate := range byGroup {
		selectedMap[candidate.Applied.PromotionID] = candidate
	}
	selected := make([]quoteCandidate, 0, len(ungrouped)+len(selectedMap))
	selected = append(selected, ungrouped...)
	for _, candidate := range selectedMap {
		selected = append(selected, candidate)
	}
	sort.SliceStable(selected, func(i, j int) bool {
		return selected[i].Stacking.Priority < selected[j].Stacking.Priority
	})
	var applied []AppliedPromotion
	for i, candidate := range selected {
		if i > 0 && !selected[i-1].Stacking.Stackable {
			rejected = append(rejected, RejectedPromotion{
				PromotionID: candidate.Applied.PromotionID,
				Code:        candidate.Applied.Code,
				Reason:      ReasonNotStackable,
			})
			continue
		}
		applied = append(applied, candidate.Applied)
	}
	return applied, rejected
}

func exclusionGroups(stacking StackingRule) []string {
	groups := nonEmptyStrings(stacking.ExclusionGroups)
	if group := strings.TrimSpace(stacking.ExclusionGroup); group != "" {
		groups = append(groups, group)
	}
	if len(groups) <= 1 {
		return groups
	}
	seen := make(map[string]struct{}, len(groups))
	out := make([]string, 0, len(groups))
	for _, group := range groups {
		key := strings.ToLower(group)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, group)
	}
	return out
}

func bestGroupConflict(byGroup map[string]quoteCandidate, groups []string) (quoteCandidate, bool) {
	var best quoteCandidate
	found := false
	for _, group := range groups {
		current, ok := byGroup[group]
		if !ok {
			continue
		}
		if !found || current.Applied.DiscountMinor > best.Applied.DiscountMinor {
			best = current
			found = true
		}
	}
	return best, found
}

func removeCandidateFromGroups(byGroup map[string]quoteCandidate, candidate quoteCandidate) {
	for group, current := range byGroup {
		if current.Applied.PromotionID == candidate.Applied.PromotionID {
			delete(byGroup, group)
		}
	}
}
