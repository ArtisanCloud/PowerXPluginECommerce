package coupon

import "sort"

// LineAmount is used for order-level discount allocations.
type LineAmount struct {
	LineID      string
	BaseMinor   int64
	DiscountCap int64
}

// LineAllocation stores allocated discount per line.
type LineAllocation struct {
	LineID        string `json:"line_id"`
	BaseMinor     int64  `json:"base_minor"`
	DiscountMinor int64  `json:"discount_minor"`
}

// AllocateDiscountByAmount allocates order discount proportionally and keeps sum exact.
func AllocateDiscountByAmount(lines []LineAmount, totalDiscountMinor int64) []LineAllocation {
	if len(lines) == 0 {
		return []LineAllocation{}
	}
	if totalDiscountMinor <= 0 {
		out := make([]LineAllocation, 0, len(lines))
		for _, line := range lines {
			out = append(out, LineAllocation{LineID: line.LineID, BaseMinor: line.BaseMinor, DiscountMinor: 0})
		}
		return out
	}
	var totalBase int64
	for _, line := range lines {
		if line.BaseMinor > 0 {
			totalBase += line.BaseMinor
		}
	}
	if totalBase <= 0 {
		return []LineAllocation{}
	}

	type remainder struct {
		idx int
		rem int64
	}
	alloc := make([]LineAllocation, 0, len(lines))
	rems := make([]remainder, 0, len(lines))
	allocated := int64(0)
	for i, line := range lines {
		if line.BaseMinor <= 0 {
			alloc = append(alloc, LineAllocation{LineID: line.LineID, BaseMinor: line.BaseMinor, DiscountMinor: 0})
			continue
		}
		numerator := totalDiscountMinor * line.BaseMinor
		value := numerator / totalBase
		if line.DiscountCap > 0 && value > line.DiscountCap {
			value = line.DiscountCap
		}
		allocated += value
		alloc = append(alloc, LineAllocation{LineID: line.LineID, BaseMinor: line.BaseMinor, DiscountMinor: value})
		rems = append(rems, remainder{idx: i, rem: numerator % totalBase})
	}

	left := totalDiscountMinor - allocated
	if left <= 0 {
		return alloc
	}
	sort.SliceStable(rems, func(i, j int) bool {
		return rems[i].rem > rems[j].rem
	})
	for _, item := range rems {
		if left <= 0 {
			break
		}
		idx := item.idx
		cap := lines[idx].DiscountCap
		if cap > 0 && alloc[idx].DiscountMinor >= cap {
			continue
		}
		alloc[idx].DiscountMinor++
		left--
	}
	return alloc
}
