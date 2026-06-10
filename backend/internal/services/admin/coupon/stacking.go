package coupon

import (
	"sort"
	"strings"
)

const (
	StackLevelItem  = "item"
	StackLevelOrder = "order"
)

// StackingMeta describes coupon stacking semantics.
type StackingMeta struct {
	AssetID        string
	Level          string
	Priority       int
	Stackable      bool
	ExclusionGroup string
}

// SortAndFilterStackable enforces ordering: item-level then order-level, then priority asc.
func SortAndFilterStackable(input []StackingMeta) (applied []StackingMeta, rejected map[string]string) {
	if len(input) == 0 {
		return []StackingMeta{}, map[string]string{}
	}
	rows := append([]StackingMeta(nil), input...)
	sort.SliceStable(rows, func(i, j int) bool {
		li := strings.TrimSpace(strings.ToLower(rows[i].Level))
		lj := strings.TrimSpace(strings.ToLower(rows[j].Level))
		if li != lj {
			if li == StackLevelItem {
				return true
			}
			if lj == StackLevelItem {
				return false
			}
		}
		if rows[i].Priority != rows[j].Priority {
			return rows[i].Priority < rows[j].Priority
		}
		return rows[i].AssetID < rows[j].AssetID
	})

	rejected = map[string]string{}
	seenExclusion := map[string]struct{}{}
	hasNonStackable := false
	for _, row := range rows {
		assetID := strings.TrimSpace(row.AssetID)
		if assetID == "" {
			continue
		}
		exclusion := strings.TrimSpace(strings.ToLower(row.ExclusionGroup))
		if exclusion != "" {
			if _, ok := seenExclusion[exclusion]; ok {
				rejected[assetID] = ReasonNotStackable
				continue
			}
		}
		if hasNonStackable {
			rejected[assetID] = ReasonNotStackable
			continue
		}
		applied = append(applied, row)
		if exclusion != "" {
			seenExclusion[exclusion] = struct{}{}
		}
		if !row.Stackable {
			hasNonStackable = true
		}
	}
	return applied, rejected
}
