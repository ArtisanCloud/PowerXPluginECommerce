package spu

import (
	"strconv"
	"strings"
)

// ImportRecord captures a single parsed row from the import file.
type ImportRecord struct {
	Row      int
	Code     string
	Name     string
	SKUCodes []string
}

// ImportFailure describes why a row cannot be imported.
type ImportFailure struct {
	Row    int    `json:"row"`
	Code   string `json:"code,omitempty"`
	Reason string `json:"reason"`
}

// ImportSummary aggregates validation output for the entire file.
type ImportSummary struct {
	TotalRows    int
	SuccessCount int
	Failures     []ImportFailure
}

// validateImportRecords enforces uniqueness and required fields, returning success/failure stats.
func validateImportRecords(records []ImportRecord) ImportSummary {
	summary := ImportSummary{TotalRows: len(records)}
	if len(records) == 0 {
		return summary
	}
	seenSPU := map[string]int{}
	seenSKU := map[string]int{}
	failures := make([]ImportFailure, 0)
	for _, record := range records {
		code := normalizeCode(record.Code)
		name := strings.TrimSpace(record.Name)
		if code == "" {
			failures = append(failures, ImportFailure{Row: record.Row, Reason: "缺少 SPU 编码"})
			continue
		}
		if name == "" {
			failures = append(failures, ImportFailure{Row: record.Row, Code: record.Code, Reason: "缺少名称"})
			continue
		}
		if prev, ok := seenSPU[code]; ok {
			failures = append(failures, ImportFailure{
				Row:    record.Row,
				Code:   record.Code,
				Reason: buildDuplicateMessage("SPU 编码", prev),
			})
			continue
		}
		if conflict := firstDuplicateSKU(record.SKUCodes, seenSKU); conflict > 0 {
			failures = append(failures, ImportFailure{
				Row:    record.Row,
				Code:   record.Code,
				Reason: buildDuplicateMessage("SKU 编码", conflict),
			})
			continue
		}
		seenSPU[code] = record.Row
		for _, sku := range record.SKUCodes {
			if skuCode := normalizeCode(sku); skuCode != "" {
				seenSKU[skuCode] = record.Row
			}
		}
		summary.SuccessCount++
	}
	summary.Failures = failures
	return summary
}

func normalizeCode(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func firstDuplicateSKU(codes []string, seen map[string]int) int {
	for _, code := range codes {
		normalized := normalizeCode(code)
		if normalized == "" {
			continue
		}
		if prev, ok := seen[normalized]; ok {
			return prev
		}
	}
	return 0
}

func buildDuplicateMessage(field string, row int) string {
	return field + "重复（已在第 " + strconv.Itoa(row) + " 行出现）"
}
