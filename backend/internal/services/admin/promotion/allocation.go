package promotion

import "sort"

type LineAmount struct {
	LineID    string
	SKUID     string
	BaseMinor int64
}

func AllocateDiscountByPrePromotionAmount(lines []LineAmount, totalDiscountMinor int64) []LineAllocation {
	out := make([]LineAllocation, 0, len(lines))
	if len(lines) == 0 {
		return out
	}
	var totalBase int64
	for _, line := range lines {
		if line.BaseMinor > 0 {
			totalBase += line.BaseMinor
		}
	}
	if totalBase <= 0 || totalDiscountMinor <= 0 {
		for _, line := range lines {
			out = append(out, LineAllocation{
				LineID: line.LineID, SKUID: line.SKUID, BaseAmountMinor: line.BaseMinor,
				AfterPromotionAmountMinor: line.BaseMinor,
			})
		}
		return out
	}
	type rem struct {
		idx int
		val int64
	}
	rems := make([]rem, 0, len(lines))
	allocated := int64(0)
	for i, line := range lines {
		value := int64(0)
		if line.BaseMinor > 0 {
			numerator := totalDiscountMinor * line.BaseMinor
			value = numerator / totalBase
			if value > line.BaseMinor {
				value = line.BaseMinor
			}
			rems = append(rems, rem{idx: i, val: numerator % totalBase})
		}
		allocated += value
		out = append(out, LineAllocation{
			LineID: line.LineID, SKUID: line.SKUID, BaseAmountMinor: line.BaseMinor,
			PromotionDiscountMinor: value, AfterPromotionAmountMinor: line.BaseMinor - value,
		})
	}
	left := totalDiscountMinor - allocated
	sort.SliceStable(rems, func(i, j int) bool {
		if rems[i].val == rems[j].val {
			return lines[rems[i].idx].BaseMinor > lines[rems[j].idx].BaseMinor
		}
		return rems[i].val > rems[j].val
	})
	for _, item := range rems {
		if left <= 0 {
			break
		}
		idx := item.idx
		if out[idx].PromotionDiscountMinor >= out[idx].BaseAmountMinor {
			continue
		}
		out[idx].PromotionDiscountMinor++
		out[idx].AfterPromotionAmountMinor--
		left--
	}
	return out
}
